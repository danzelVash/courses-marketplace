package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/danzelVash/courses-marketplace"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func (h *Handler) oauthVK(ctx *gin.Context) {
	stateTemp := ctx.Request.URL.Query().Get("state")
	code := ctx.Request.URL.Query().Get("code")

	if !h.services.ValidateParamsFromRedirect(stateTemp, code) {
		newRespErr(ctx, http.StatusBadRequest, courses.BadVkInteraction.Error())
		return
	}

	url := h.services.GenerateUrlForVKHandshake(code)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		newRespErr(ctx, http.StatusBadGateway, courses.BadVkInteraction.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		newRespErr(ctx, http.StatusBadRequest, courses.BadVkInteraction.Error())
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Error("memory leak error while closing response body")
		}
	}(resp.Body)

	url, err = h.services.GenerateUrlForVkApi(resp.Body)
	if err != nil {
		newRespErr(ctx, http.StatusBadGateway, courses.BadVkInteraction.Error())
	}

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		newRespErr(ctx, http.StatusBadGateway, courses.BadVkInteraction.Error())
		return
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		newRespErr(ctx, http.StatusBadRequest, courses.BadVkInteraction.Error())
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Error("unclosed response.Body in oauthVK")
		}
	}(resp.Body)

	SID, err := h.services.AuthorizeVkUser(resp.Body)
	if err != nil {
		newRespErr(ctx, http.StatusBadRequest, courses.BadVkInteraction.Error())
		return
	}

	ctx.SetCookie("session_id", SID, 60*60*24*14, "/", "localhost", true, false)

	ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/api/", viper.GetString("domain")))
}
