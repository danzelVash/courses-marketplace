package m3u8_generator

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

/*
ffmpeg -i input.mp4 \
-map 0:v:0 -map 0:a:0 -c:v libx264 -b:v 6000k -c:a aac -b:a 128k -preset veryfast -vf scale=1920:-2 -f hls -hls_list_size 0 -hls_time 10 -hls_segment_filename "1080p_%03d.ts" 1080p.m3u8 \
-map 0:v:0 -map 0:a:0 -c:v libx264 -b:v 3000k -c:a aac -b:a 128k -preset veryfast -vf scale=1280:-2 -f hls -hls_list_size 0 -hls_time 10 -hls_segment_filename "720p_%03d.ts" 720p.m3u8 \
-map 0:v:0 -map 0:a:0 -c:v libx264 -b:v 1000k -c:a aac -b:a 128k -preset veryfast -vf scale=854:-2 -f hls -hls_list_size 0 -hls_time 10 -hls_segment_filename "480p_%03d.ts" 480p.m3u8
*/

// TODO add logger into this package

const (
	hlsFragmentDuration = "10"
	globTmpDir          = "internal/libraries/m3u8_generator/tmp"
	resolutionsForCmd   = "1920:1280:854"
	resNames            = "1080:720:480"
	bitrateP1080        = "6000"
	bitrateP720         = "3000"
	bitrateP480         = "1000"
)

var (
	FailToMakeMasterPlaylist = errors.New("fail to make master play-list")
)

type playListParams struct {
	resolutionForCmd string
	playListDir      string
	pathToVideo      string
	bitrate          string
	fragmentName     string
}

type responseFromPlayListGenerator struct {
	err                 error
	resolution          string
	resolutionFileNames []string
}

type AnswerToCaller struct {
	Err      error
	FileName string
	Folder   string
	Data     *os.File
}

func makePlaylist(ctx context.Context, resChan chan<- *responseFromPlayListGenerator, params *playListParams, wg *sync.WaitGroup) {
	defer wg.Done()
	//if _, err := os.Stat(filepath.Join(pathToCurDir, params.playListDir)); os.IsNotExist(err) {
	//	err := os.MkdirAll(filepath.Join(pathToCurDir, params.playListDir), os.ModePerm)
	//	if err != nil {
	//		resChan <- &responseFromPlayListGenerator{err: err}
	//		return
	//	}
	//}
	cmd := exec.Command(
		"ffmpeg",
		"-i", params.pathToVideo,
		"-map", "0:v:0",
		"-map", "0:a:0",
		"-c:v", "libx264",
		"-b:v", params.bitrate+"k",
		"-c:a", "aac",
		"-b:a", "128k",
		"-preset", "veryfast",
		"-vf", fmt.Sprintf("scale=%s:-2", params.resolutionForCmd),
		"-f", "hls",
		"-hls_list_size", "0",
		"-hls_time", hlsFragmentDuration,
		"-hls_segment_filename", filepath.Join(params.playListDir, fmt.Sprintf("%sp_", params.fragmentName)+"%03d.ts"),
		filepath.Join(params.playListDir, fmt.Sprintf("%s.m3u8", params.fragmentName)),
	)

	err := cmd.Run()
	if err != nil {
		fmt.Println("error", err)
		resChan <- &responseFromPlayListGenerator{err: err}
		return
	}

	files, err := os.ReadDir(params.playListDir)
	if err != nil {
		fmt.Println("error 2", err)
		resChan <- &responseFromPlayListGenerator{err: err}
		return
	}

	answer := &responseFromPlayListGenerator{resolution: params.fragmentName, resolutionFileNames: make([]string, 0, len(files))}

	for _, file := range files {
		if !file.IsDir() {
			answer.resolutionFileNames = append(answer.resolutionFileNames, filepath.Join(params.playListDir, file.Name()))
		}
	}

	fmt.Println(answer.resolutionFileNames)

	answer.err = nil

	fmt.Println("all was done, fileNames was sent to channel")
	resChan <- answer
}

func sendIoReaderToChan(out chan<- *AnswerToCaller, filePaths []string, playListDir string) {
	for _, path := range filePaths {

		// файл закрывается после отправки в object storage
		f, err := os.Open(path)
		if err != nil {
			out <- &AnswerToCaller{Err: err}
			continue
		}
		out <- &AnswerToCaller{
			Err:      nil,
			FileName: filepath.Base(path),
			Folder:   playListDir,
			Data:     f,
		}
	}
}

// TODO clean code and add logging

func CutVideoToHLSFragments(ctx context.Context, videoName, playListDir string, outChan chan<- *AnswerToCaller) {
	fmt.Println("CutVideoToHLSFragments was called")
	defer close(outChan)
	resultDir := filepath.Join(globTmpDir, playListDir)
	if _, err := os.Stat(resultDir); os.IsNotExist(err) {
		if err := os.MkdirAll(resultDir, os.ModePerm); err != nil {
			outChan <- &AnswerToCaller{Err: err}
			return
		}
	}

	resolutionArr := strings.Split(resolutionsForCmd, ":")
	fragmentNames := strings.Split(resNames, ":")
	bitrates := map[string]string{
		resolutionArr[0]: bitrateP1080,
		resolutionArr[1]: bitrateP720,
		resolutionArr[2]: bitrateP480,
	}

	wg := &sync.WaitGroup{}

	resChan := make(chan *responseFromPlayListGenerator, 3)
	for i, resolution := range resolutionArr {
		resDir := filepath.Join(globTmpDir, playListDir, fragmentNames[i])
		if _, err := os.Stat(resDir); os.IsNotExist(err) {
			if err := os.MkdirAll(resDir, os.ModePerm); err != nil {
				fmt.Println("error while creating directory", err.Error())
				outChan <- &AnswerToCaller{Err: err}
				return
			}
		}
		plParams := &playListParams{
			resolutionForCmd: resolution,
			playListDir:      resDir,
			pathToVideo:      filepath.Join(globTmpDir, videoName),
			bitrate:          bitrates[resolution],
			fragmentName:     fragmentNames[i],
		}

		wg.Add(1)
		go makePlaylist(ctx, resChan, plParams, wg)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	success := 0

	for res := range resChan {
		if res.err != nil {
			fmt.Println(res.err)
			success++
		} else {
			success++
			go sendIoReaderToChan(outChan, res.resolutionFileNames, playListDir)
		}

	}

	if success == len(resolutionArr) {
		master := "#EXTM3U\n" +
			"#EXT-X-VERSION:3\n" +
			"#EXT-X-STREAM-INF:BANDWIDTH=6000000,RESOLUTION=1920x1080,NAME=\"1080\"\n1080p.m3u8\n" +
			"#EXT-X-STREAM-INF:BANDWIDTH=3000000,RESOLUTION=1280x1080,NAME=\"720\"\n720p.m3u8\n" +
			"#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=854x480,NAME=\"480\"\n480p.m3u8"
		fileName := fmt.Sprintf("%s.m3u8", playListDir)
		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)

		if err != nil {
			outChan <- &AnswerToCaller{Err: errors.WithMessage(FailToMakeMasterPlaylist, err.Error())}
			return
		}

		_, err = f.WriteString(master)
		if err != nil {
			outChan <- &AnswerToCaller{Err: errors.WithMessage(FailToMakeMasterPlaylist, err.Error())}
			return
		}

		outChan <- &AnswerToCaller{Err: nil, FileName: fileName, Folder: playListDir, Data: f}
	} else {
		outChan <- &AnswerToCaller{Err: errors.WithMessage(FailToMakeMasterPlaylist, "problem with goroutines while cutting one of resolutions")}
	}
}
