package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	courses "mom"
	"mom/internal/libraries/template_parser"
	"net/http"
	"strconv"
)

func (h *Handler) signUpPage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "public/sign_up_form.html",
		Vars: struct {
			Url    string
			Domain string
		}{
			Url:    h.services.Authorization.GenerateLinkForOauth(),
			Domain: viper.GetString("domain"),
		},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)
}

func (h *Handler) signInPage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "public/sign_in_form.html",
		Vars: struct {
			Url    string
			Domain string
		}{
			Url:    h.services.Authorization.GenerateLinkForOauth(),
			Domain: viper.GetString("domain"),
		},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)
}

func (h *Handler) signOutPage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "public/sign_out_form.html",
		Vars: struct {
			Domain string
		}{
			Domain: viper.GetString("domain"),
		},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)
}

func (h *Handler) forgotPasswordPage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "public/sign_in_forgot.html",
		Vars: struct {
			Domain string
		}{
			Domain: viper.GetString("domain"),
		},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)
}

func (h *Handler) PasswordUpdatePage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "public/update_password.html",
		Vars: struct {
			Domain string
		}{
			Domain: viper.GetString("domain"),
		},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)
}

func (h *Handler) signUp(ctx *gin.Context) {
	user := courses.User{}

	if err := ctx.BindJSON(&user); err != nil {
		h.logger.Errorf("error while binding json: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	user.Vk = false

	_, err := h.services.Authorization.CreateUser(user)

	switch err {
	case courses.AlreadyRegistered:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist": true,
		})
	case nil:
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/auth/sign-in/", viper.GetString("domain")))
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"exist": false,
		})
	}
}

func (h *Handler) signIn(ctx *gin.Context) {
	input := struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}

	if err := ctx.BindJSON(&input); err != nil {
		h.logger.Errorf("error while binding json: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	SID, err := h.services.Authorization.CreateSession(input.Login, input.Password)

	switch err {
	case courses.ErrNoRows:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist":      false,
			"right_pass": false,
		})
	case courses.BadPassword:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist":      true,
			"right_pass": false,
		})
	case nil:
		ctx.SetCookie("session_id", SID, 60*60*24*14, "/", viper.GetString("set_cookie_domain"), true, false)
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/api/", viper.GetString("domain")))
	default:
		ctx.JSON(http.StatusBadGateway, gin.H{
			"exist":      false,
			"right_pass": false,
		})
	}
}

func (h *Handler) sendCodeForPasswordUpdate(ctx *gin.Context) {
	email := struct {
		Email string `json:"email" binding:"required"`
	}{}

	if err := ctx.BindJSON(&email); err != nil {
		h.logger.Errorf("error while binding json: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, "incorrect email")
	}

	err := h.services.Authorization.CheckEmailAndSendCode(email.Email)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, gin.H{
			"valid": true,
			"exist": true,
		})
		return
	case courses.ErrNoRows:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist": false,
			"valid": false,
		})
		return
	case courses.BadEmail:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist": false,
			"valid": false,
		})
	default:
		ctx.JSON(http.StatusBadGateway, gin.H{
			"exist": false,
			"valid": false,
		})
	}
}

func (h *Handler) PasswordUpdate(ctx *gin.Context) {
	code := struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}{}

	if err := ctx.BindJSON(&code); err != nil {
		h.logger.Errorf("error while binding json: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"correct":          false,
			"password_updated": false,
		})
		return
	}

	numCode, err := strconv.Atoi(code.Code)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"valid_code":       false,
			"password_updated": false,
		})
		return
	}

	if err := h.services.CheckCode(code.Email, numCode); err == courses.BadCode {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"valid_code":       false,
			"password_updated": false,
		})
		return
	}

	if err := h.services.UpdatePassword(code.Email); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"valid_code":       false,
			"password_updated": false,
		})
		return
	}

	ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/auth/sign-in/)", viper.GetString("domain")))
	//switch err {
	//case nil:
	//	ctx.SetCookie("update_token", token, 60*5, "/", "localhost", true, false)
	//	ctx.Redirect(http.StatusSeeOther, "http://localhost:8080/sign-in/forgot-pass/update/")
	//case courses.BadCode:
	//	ctx.JSON(http.StatusBadRequest, gin.H{
	//		"correct": false,
	//	})
	//default:
	//	ctx.JSON(http.StatusInternalServerError, gin.H{
	//		"correct": false,
	//	})
	//}
}

//func (h *Handler) PasswordUpdate(ctx *gin.Context) {
//	token, err := ctx.Cookie("update_token")
//	if err != nil {
//		ctx.JSON(http.StatusUnauthorized, gin.H{})
//		return
//	}
//
//	email, err := h.services.Authorization.ParseToken(token)
//	if err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"valid_token": "false",
//		})
//		return
//	}
//
//	if err := h.services.Authorization.UpdatePassword(email); err != nil {
//
//	}
//}

func (h *Handler) signOut(ctx *gin.Context) {
	sessionId, err := ctx.Cookie("session_id")
	if err != nil {
		h.logger.Info("user trying to sign out without session")
		newRespErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Authorization.DeleteSession(sessionId); err != nil {
		h.logger.Errorf("error while deleting session, responsed with status 500: %s", err.Error())
		newRespErr(ctx, http.StatusInternalServerError, "error while signing out")
	}

	ctx.SetCookie("session_id", sessionId, -1, "/", viper.GetString("set_cookie_domain"), true, false)
	ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/", viper.GetString("domain")))
}
