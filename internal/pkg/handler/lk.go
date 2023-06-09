package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"mom/internal/libraries/aws_sdk"
	"mom/internal/libraries/template_parser"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	yandexBucketName = "test-videos-hls"
)

func (h *Handler) lk(ctx *gin.Context) {
	_, ok := ctx.Get(userCtx)
	if !ok {
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/auth/sign-in/", viper.GetString("domain")))
	}

	params := template_parser.TemplateParams{
		TemplateName: "user_api/test_lk.html",
		Vars:         struct{}{},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)

	//ctx.HTML(http.StatusOK, "test_lk.html", gin.H{})
}

func (h *Handler) videoHandler(ctx *gin.Context) {
	//if _, ok := ctx.Get(userCtx); ok {
	//	newRespErr(ctx, http.StatusForbidden, "i don`t know what you want")
	//	return
	//}
	dir := ctx.Param("id")
	fileName := ctx.Param("filename")
	if _, err := strconv.Atoi(dir); err == nil && (strings.HasSuffix(fileName, ".m3u8") || strings.HasSuffix(fileName, ".ts")) {
		filePath := filepath.Join(dir, fileName)

		bytes, err := aws_sdk.GetObjectFromYandexCloud(yandexBucketName, filePath)
		if err != nil {
			h.logger.Error(err)
			newRespErr(ctx, http.StatusBadGateway, "error with s3 storage")
			return
		}

		//ctx.Stream(func(w io.Writer) bool {
		//	if _, err := fmt.Fprint(w, bytes); err != nil {
		//		return false
		//	}
		//	return true
		//})
		ctx.Header("Content-Disposition", "inline")
		if strings.HasSuffix(fileName, ".m3u8") {
			ctx.Data(http.StatusOK, "application/vnd.apple.mpegurl", bytes)
		} else {
			ctx.Data(http.StatusOK, "video/MP2T", bytes)
		}

		//defer func(writer gin.ResponseWriter) {
		//	if err := recover(); err != nil {
		//		if ne, ok := err.(net.Error); ok && ne.Timeout() {
		//			h.logger.Errorf("timeout error occured while trying to send video fragment: %v", err)
		//		} else if !writer.Written() {
		//			h.logger.Errorf("data was not written to client`s thread: %v", err)
		//		}
		//	}
		//}(ctx.Writer)

		h.logger.Infof("sent video fragment %s", fileName)
	} else {
		newRespErr(ctx, http.StatusNotFound, "invalid filepath:"+ctx.Request.URL.String())
	}

}

// id ключа
//YCAJEmZbeGxDArIoF33HOoFdK
//secret
//YCNEKkLF_UZbDqbnryYwUX_AttbI107yOOfvzAZA
