package service

import (
	"context"
	"github.com/spf13/viper"
	"math/rand"
	courses "mom"
	"mom/internal/libraries/aws_sdk"
	"mom/internal/libraries/logging"
	"mom/internal/libraries/m3u8_generator"
	"mom/internal/libraries/smtp"
	"mom/internal/pkg/repository"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	codeLength         = 6
	symbols            = "0123456789"
	videoLessonsBucket = "test-videos-hls"
)

type AdminService struct {
	repos  repository.Administration
	logger *logging.Logger
}

func NewAdminService(repos repository.Administration, logger *logging.Logger) *AdminService {
	return &AdminService{
		repos:  repos,
		logger: logger,
	}
}

func (adm *AdminService) SendAuthCode(admin courses.Administrator) error {
	validAdmin := courses.Administrator{
		Login:    os.Getenv("ADMIN_LOGIN"),
		Password: os.Getenv("ADMIN_PASSWORD"),
	}

	if validAdmin.Login != admin.Login || validAdmin.Password != admin.Password {
		return courses.BadAdminAuth
	}

	code := generateAuthCode()

	emailParams := smtp.EmailParams{
		TemplateName: "templates/email_templates/message_to_admin_mail.html",
		TemplateVars: struct {
			Code int
			Url  string
		}{
			Code: code,
			Url:  viper.GetString("domain"),
		},
		Destination: os.Getenv("ADMIN_MAIL"),
		Subject:     "Вход в панель администратора",
	}

	if err := smtp.SendEmail(emailParams); err != nil {
		adm.logger.Errorf("error while sending admin email: %s", err.Error())
		return err
	}

	adm.repos.AddCode(emailParams.Destination, code)

	return nil
}

func (adm *AdminService) CreateAdminSession(inputCode int) (string, error) {
	code, ok := adm.repos.GetCode(os.Getenv("ADMIN_MAIL"))

	if !ok {
		return "", courses.NoActiveCodesForEmail
	}

	if code != inputCode {
		return "", courses.IncorrectAuthCode
	}

	if err := adm.repos.DeleteCode(os.Getenv("ADMIN_MAIL")); err != nil {
		adm.logger.Error("error while deleting code from storage")
	}
	SID := generateRandomString(sessLen)
	expiredDate := time.Now().AddDate(0, 0, 1).Round(time.Hour)
	if _, err := adm.repos.CreateSession(1, SID, expiredDate); err != nil {
		adm.logger.Errorf("error while creating session in auth_pstgrs %s", err.Error())
		return "", err
	}

	return SID, nil

}

func (adm *AdminService) CheckAdminSession(sessionId string) (int, error) {
	userId, err := adm.repos.CheckSession(sessionId)

	if err != nil {
		return 0, err
	}

	return userId, err
}

func (adm *AdminService) SelectAllUsers() ([]courses.User, error) {
	return adm.repos.SelectAllUsers()
}

func (adm *AdminService) CreateVideoLessonOnBucket(ctx context.Context, bytesChan <-chan *m3u8_generator.AnswerToCaller) {
	for {
		select {
		case <-ctx.Done():
			adm.logger.Infof("context done: %s", ctx.Err())
			return
		case answer, open := <-bytesChan:
			if open {
				switch answer.Err {
				case nil:
					err := aws_sdk.PutObjectInBucket(videoLessonsBucket, filepath.Join(answer.Folder, answer.FileName), answer.Data)
					if err != nil {
						adm.logger.Errorf("error while Putting object to yandex cloud: %s", err.Error())
					}
				case m3u8_generator.FailToMakeMasterPlaylist:
					adm.logger.Errorf("error creating master play-list: %s", answer.Err)
				default:
					adm.logger.Errorf("unknown error occured while reading from answer channel: %s", answer.Err)
				}
			} else {
				adm.logger.Info("channel was closen and video lesson must be uploaded to object storage")
				return
			}
		}
	}
}

func generateAuthCode() int {
	rand.Seed(time.Now().UnixNano())
	codeStr := ""
	for i := 0; i < codeLength; i++ {
		if i != 0 {
			codeStr += string(symbols[rand.Intn(len(symbols))])
			continue
		}
		codeStr += string(symbols[1:][rand.Intn(len(symbols)-1)])
	}

	code, _ := strconv.Atoi(codeStr)

	return code
}
