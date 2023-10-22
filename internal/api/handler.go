package api

import (
	"github.com/danzelVash/courses-marketplace/internal/service"
	"github.com/danzelVash/courses-marketplace/pkg/logging"
	"github.com/danzelVash/courses-marketplace/pkg/template_parser"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"net/http"
)

type Handler struct {
	services *service.Service
	logger   *logging.Logger
}

func NewHandler(services *service.Service, logger *logging.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) InitRouters() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	//router.MaxMultipartMemory = 5 << 30
	//router.Use(h.accessLogMiddleware())

	//router.Static("/css", "templates/css")
	//router.Static("/img", "templates/img")
	//router.Static("/js", "templates/js")

	router.GET("/", func(ctx *gin.Context) {
		params := template_parser.TemplateParams{
			TemplateName: "public/index.html",
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
	})

	auth := router.Group("/auth", h.CheckAuthGroup)
	{
		auth.GET("/oauthVK/", h.oauthVK)

		signIn := auth.Group("/sign-in")
		{
			signIn.GET("/", h.signInPage)
			signIn.POST("/", h.signIn)

			forgotPass := signIn.Group("/forgot-pass")
			{
				forgotPass.GET("/", h.forgotPasswordPage)
				forgotPass.POST("/", h.PasswordUpdate)
				forgotPass.POST("/send-code/", h.sendCodeForPasswordUpdate)
			}

		}

		auth.GET("/sign-up/", h.signUpPage)
		auth.POST("/sign-up/", h.signUp)

		auth.GET("sign-out/", h.signOutPage)
		auth.POST("/sign-out/", h.signOut)
	}

	api := router.Group("/api", h.IdentifyUser)
	{
		api.GET("/videos/:id/:filename", h.videoHandler)
		api.GET("/", h.lk)

		//courses := api.Group("/courses")
		//{
		//	courses.POST("/:id", h.buyCourse)
		//	courses.GET("/:id", h.getCourseById)
		//
		//	items := courses.Group(":id/items")
		//	{
		//		items.GET("/", h.getAllItems)
		//		items.GET("/:id", h.getItemById)
		//	}
		//}
	}

	admin := router.Group("/admin")
	{
		adminAuth := admin.Group("/auth", h.CheckAdminAuth)
		{
			adminAuth.GET("/", h.adminAuthPage)
			adminAuth.POST("/", h.authAdmin)

			adminAuth.GET("/code/", h.checkAdminCodePage)
			adminAuth.POST("/code/", h.checkAdminCode)
		}

		adminPanel := admin.Group("/panel", h.IdentifyAdmin)
		{
			adminPanel.GET("/", h.adminPanelPage)

			users := adminPanel.Group("/users")
			{
				users.GET("/", h.panelUsersListPage)

				users.GET("/upload/", h.panelUserUploadPage)

				users.GET("/:id/", h.panelUserPage)
				users.PUT("/:id/")

				users.DELETE("/:id/")
			}

			videoLessons := adminPanel.Group("/video-lessons")
			{
				videoLessons.GET("/", h.videoLessonsPage)
				videoLessons.GET("upload/", h.videoLessonsUploadPage)
				videoLessons.POST("upload/", h.videoLessonUpload)
			}
		}
	}

	return router
}
