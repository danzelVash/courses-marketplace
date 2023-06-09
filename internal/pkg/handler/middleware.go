package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

const (
	userCtx  = "userId"
	adminCtx = "adminId"
)

func (h *Handler) IdentifyUser(ctx *gin.Context) {
	cookie, err := ctx.Cookie("session_id")
	if err != nil {
		newRespErr(ctx, http.StatusUnauthorized, "have no session")
		return
	}

	userId, err := h.services.Authorization.CheckSession(cookie)

	if err != nil {
		newRespErr(ctx, http.StatusUnauthorized, "there is no active session like your")
		return
	}

	ctx.Set(userCtx, userId)
}

func (h *Handler) CheckAuthGroup(ctx *gin.Context) {
	if sessionId, err := ctx.Cookie("session_id"); err == nil {
		if _, err := h.services.Authorization.CheckSession(sessionId); err == nil {
			ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/api/", viper.GetString("domain")))
			return
		}
	}
}

func (h *Handler) IdentifyAdmin(ctx *gin.Context) {
	cookie, err := ctx.Cookie("admin_session_id")

	if err != nil {
		newRespErr(ctx, http.StatusUnauthorized, "have no session")
		return
	}

	userId, err := h.services.Administration.CheckAdminSession(cookie)

	if err != nil {
		newRespErr(ctx, http.StatusUnauthorized, "there is no active session like your")
		return
	}

	ctx.Set(adminCtx, userId)
}

func (h *Handler) CheckAdminAuth(ctx *gin.Context) {
	if cookie, err := ctx.Cookie("admin_session_id"); err == nil {
		if _, err = h.services.Administration.CheckAdminSession(cookie); err == nil {
			ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/admin/panel/", viper.GetString("domain")))
			return
		}
	}
}

func (h *Handler) accessLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		defer func() {
			dur := time.Since(start)
			h.logger.Infof("accessLogMiddleware on %s was carried out in %v", ctx.Request.URL.String(), dur)
		}()
		defer func() {
			if err := recover(); err != nil {
				h.logger.Errorf("Panic occurred: %s", err)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		defer ctx.Next()
	}

}

func (h *Handler) PanicMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				h.logger.Errorf("Panic occurred: %s", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	}
}
