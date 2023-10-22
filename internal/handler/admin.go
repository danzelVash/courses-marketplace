package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"

	"github.com/danzelVash/courses-marketplace"
	m3u8 "github.com/danzelVash/courses-marketplace/pkg/m3u8_generator"
	"github.com/danzelVash/courses-marketplace/pkg/template_parser"
	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-unidecode"
	"github.com/spf13/viper"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	m3u8TmpDir            = "internal/pkg/m3u8_generator/tmp"
	additionalFilesTmpDir = "internal/pkg/handler/tmp"
)

func (h *Handler) adminAuthPage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "admin/auth.html",
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

func (h *Handler) authAdmin(ctx *gin.Context) {
	admin := courses.Administrator{}

	if err := ctx.BindJSON(&admin); err != nil {
		h.logger.Errorf("error while binding json: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	err := h.services.SendAuthCode(admin)

	switch err {
	case courses.BadAdminAuth:
		newRespErr(ctx, http.StatusUnauthorized, "you are not the admin")
		h.logger.Infof("user with bad auth trying to login as admin")
	case courses.ErrorSendingMail:
		newRespErr(ctx, http.StatusBadGateway, "server error")
		h.logger.Errorf("error while sending email: %s", err.Error())
	case nil:
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/admin/auth/code/", viper.GetString("domain")))
		h.logger.Info("user was login as admin, need to check email")
	default:
		newRespErr(ctx, http.StatusInternalServerError, err.Error())
		h.logger.Errorf("unknown error, responsed with status 500 : %s", err.Error())
	}
}

func (h *Handler) checkAdminCodePage(ctx *gin.Context) {
	params := template_parser.TemplateParams{
		TemplateName: "admin/auth_code.html",
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

func (h *Handler) checkAdminCode(ctx *gin.Context) {
	code := struct {
		Code string `json:"code" binding:"required"`
	}{}

	if err := ctx.BindJSON(&code); err != nil {
		h.logger.Errorf("error while binding json: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, "invalid code")
		return
	}

	numCode, err := strconv.Atoi(code.Code)
	if err != nil {
		h.logger.Infof("invalid auth code while trying to login as admin: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, "invalid auth code")
		return
	}

	SID, err := h.services.CreateAdminSession(numCode)

	switch err {
	case courses.NoActiveCodesForEmail:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist":   false,
			"correct": false,
		})
	case courses.IncorrectAuthCode:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"exist":   true,
			"correct": false,
		})
	case nil:
		ctx.SetCookie("admin_session_id", SID, 60*60*24, "/admin/", viper.GetString("set_cookie_domain"), true, false)
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/admin/panel/", viper.GetString("domain")))
	default:
		h.logger.Errorf("unknown error occured responsed with status 500: %s", err.Error())
		newRespErr(ctx, http.StatusInternalServerError, "error in server logic")
	}
}

func (h *Handler) adminPanelPage(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
		return
	}

	templateNames := []string{"admin/left_bar.html", "admin/footer.html"}
	templates, err := template_parser.GetTemplates(templateNames)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "error while executing template")
		return
	}

	params := template_parser.TemplateParams{
		TemplateName: "admin/panel.html",
		Vars: struct {
			LeftBar template.HTML
			Footer  template.HTML
			Domain  string
		}{
			LeftBar: template.HTML(templates[0]),
			Footer:  template.HTML(templates[1]),
			Domain:  viper.GetString("domain"),
		},
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)

	//ctx.HTML(http.StatusOK, "panel.html", gin.H{})
}

func (h *Handler) panelUsersListPage(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
		return
	}

	templateNames := []string{"admin/left_bar.html", "admin/footer.html"}
	templates, err := template_parser.GetTemplates(templateNames)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "error while executing template")
		return
	}

	params := template_parser.TemplateParams{
		TemplateName: "admin/users.html",
		Vars: struct {
			LeftBar template.HTML
			Footer  template.HTML
			Users   []courses.User
			Domain  string
		}{
			LeftBar: template.HTML(templates[0]),
			Footer:  template.HTML(templates[1]),
			Domain:  viper.GetString("domain"),
		},
	}

	users, err := h.services.SelectAllUsers()
	if err != nil {
		data, err := template_parser.TemplateParser(params)
		if err != nil {
			h.logger.Error(err)
			newRespErr(ctx, http.StatusInternalServerError, "internal server error")
			return
		}
		ctx.Data(http.StatusOK, "text/html", data)
		return
	}

	params.Vars = struct {
		LeftBar template.HTML
		Footer  template.HTML
		Users   []courses.User
		Domain  string
	}{
		LeftBar: template.HTML(templates[0]),
		Footer:  template.HTML(templates[1]),
		Users:   users,
		Domain:  viper.GetString("domain"),
	}

	data, err := template_parser.TemplateParser(params)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx.Data(http.StatusOK, "text/html", data)
}

func (h *Handler) panelUserPage(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
		return
	}

	userId := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{
		"user_id": userId,
	})
}

func (h *Handler) panelUserUploadPage(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
		return
	}

	templateNames := []string{"admin/left_bar.html", "admin/footer.html"}
	templates, err := template_parser.GetTemplates(templateNames)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "error while executing template")
		return
	}

	params := template_parser.TemplateParams{
		TemplateName: "admin/user_upload.html",
		Vars: struct {
			LeftBar template.HTML
			Footer  template.HTML
			Domain  string
		}{
			LeftBar: template.HTML(templates[0]),
			Footer:  template.HTML(templates[1]),
			Domain:  viper.GetString("domain"),
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

func (h *Handler) videoLessonsPage(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
		return
	}

	templateNames := []string{"admin/left_bar.html", "admin/footer.html"}
	templates, err := template_parser.GetTemplates(templateNames)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "error while executing template")
		return
	}

	params := template_parser.TemplateParams{
		TemplateName: "admin/video_lessons.html",
		Vars: struct {
			LeftBar template.HTML
			Footer  template.HTML
			Domain  string
		}{
			LeftBar: template.HTML(templates[0]),
			Footer:  template.HTML(templates[1]),
			Domain:  viper.GetString("domain"),
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

func (h *Handler) videoLessonsUploadPage(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
		return
	}

	templateNames := []string{"admin/left_bar.html", "admin/footer.html"}
	templates, err := template_parser.GetTemplates(templateNames)
	if err != nil {
		h.logger.Error(err)
		newRespErr(ctx, http.StatusInternalServerError, "error while executing template")
		return
	}

	params := template_parser.TemplateParams{
		TemplateName: "admin/video_lessons_upload.html",
		Vars: struct {
			LeftBar template.HTML
			Footer  template.HTML
			Domain  string
		}{
			LeftBar: template.HTML(templates[0]),
			Footer:  template.HTML(templates[1]),
			Domain:  viper.GetString("domain"),
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

//var upgrader = websocket.Upgrader{
//	ReadBufferSize:  1024,
//	WriteBufferSize: 1024,
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}

func (h *Handler) videoLessonUpload(ctx *gin.Context) {
	_, ok := ctx.Get(adminCtx)
	if !ok {
		h.logger.Info("user trying to enter as admin with no admin_session")
		newRespErr(ctx, http.StatusForbidden, "no admin session")
	}

	//newCtx, cancel := context.WithTimeout(ctx.Request.Context(), 30*time.Second)
	//defer cancel()
	//ctx.Request = ctx.Request.WithContext(newCtx)

	start := time.Now()

	var videoLesson courses.VideoLesson
	multipartReader, err := ctx.Request.MultipartReader()
	if err != nil {
		h.logger.Errorf("error while creating new multipart reader: %s", err.Error())
		newRespErr(ctx, http.StatusBadRequest, "incorrect request")
		return
	}

	textFields := map[string]struct{}{
		"title":                {},
		"description":          {},
		"price":                {},
		"video_name":           {},
		"additional_filenames": {},
	}

	fileFields := map[string]struct{}{
		"video": {},
	}

	for {
		chunk, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			h.logger.Errorf("unknown error occured while get next part of multipart reader: %s", err.Error())
			newRespErr(ctx, http.StatusInternalServerError, "internal server error")
			return
		}

		if _, ok := fileFields[chunk.FormName()]; ok {
			if fileName := chunk.FileName(); fileName != "" {
				fileName = unidecode.Unidecode(fileName)
				fmt.Println(fileName)

				file, err := os.OpenFile(filepath.Join(m3u8TmpDir, fileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
				if err != nil {
					h.logger.Errorf("error while os.OpenFile(): %s", err.Error())
					newRespErr(ctx, http.StatusInternalServerError, "internal server error")
					return
				}

				written, err := io.Copy(file, chunk)
				if err != nil {
					h.logger.Errorf("error while io.Copy while copying %s: %s", fileName, err.Error())
					newRespErr(ctx, http.StatusInternalServerError, "internal server error")
					return
				}
				h.logger.Infof("%d bytes was writen into %s", written, fileName)

				if err := file.Close(); err != nil {
					h.logger.Infof("memory leak: error closing file: %s", err.Error())
				}

				playListDir := fileName[:utf8.RuneCountInString(filepath.Ext(fileName))+1]
				bytesChan := make(chan *m3u8.AnswerToCaller, 100)
				go m3u8.CutVideoToHLSFragments(ctx, fileName, playListDir, bytesChan)
				go h.services.Administration.CreateVideoLessonOnBucket(ctx, bytesChan)
			}
		} else if _, ok := textFields[chunk.FormName()]; ok {
			switch chunk.FormName() {
			case "title":
				_, title, err := writeMultipartPartToString(chunk)
				if err != nil {
					h.logger.Errorf("error copying to buf: %s", err.Error())
					newRespErr(ctx, http.StatusInternalServerError, "internal server error")
					return
				}
				videoLesson.Title = title
			case "description":
				_, desc, err := writeMultipartPartToString(chunk)
				if err != nil {
					h.logger.Errorf("error copying to buf: %s", err.Error())
					newRespErr(ctx, http.StatusInternalServerError, "internal server error")
					return
				}
				videoLesson.Description = desc
			case "price":
				_, strPrice, err := writeMultipartPartToString(chunk)
				if err != nil {
					h.logger.Errorf("error copying to buf: %s", err.Error())
				}
				price, err := strconv.Atoi(strPrice)
				if err != nil {
					h.logger.Errorf("price is not number: %s", err.Error())
					newRespErr(ctx, http.StatusBadRequest, "invalid request")
					return
				}
				videoLesson.Price = price
			default:
				h.logger.Infof("unknown chunk.FormName() = %s", chunk.FormName())
			}
		}
	}

	files := ctx.Request.MultipartForm.File["additional_files"]
	for _, fileHeader := range files {
		if err := ctx.SaveUploadedFile(fileHeader, filepath.Join(additionalFilesTmpDir, fileHeader.Filename)); err != nil {
			h.logger.Errorf("error while saving additional file to ssd: %s", err.Error())
		}
	}

	end := time.Since(start)
	h.logger.Infof("(FormData): all was done at :%v", end)
}

func writeMultipartPartToString(chunk *multipart.Part) (int64, string, error) {
	buf := new(bytes.Buffer)
	written, err := io.Copy(buf, chunk)
	if err != nil {
		return 0, "", err
	}
	return written, buf.String(), nil
}

//func SaveImage(data []byte, filename string) error {
//	img, _, err := image.Decode(bytes.NewReader(data))
//	if err != nil {
//		return err
//	}
//
//	file, err := os.Create("internal/pkg/handler/saved/" + filename)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	err = jpeg.Encode(file, img, nil)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
