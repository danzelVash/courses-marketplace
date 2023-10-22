package api

import (
	"github.com/gin-gonic/gin"
)

type respErr struct {
	Msg string `json:"msg"`
}

func newRespErr(ctx *gin.Context, statusCode int, msg string) {
	ctx.AbortWithStatusJSON(statusCode, respErr{Msg: msg})
}
