package auth

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	handler := NewHandler(NewService(deps.DB, deps.JWT))

	auth := rg.Group("/auth")
	auth.POST("/register", deps.AuthRateMW, handler.Register)
	auth.POST("/login", deps.AuthRateMW, handler.Login)
	auth.POST("/password/forgot", deps.AuthRateMW, handler.ForgotPassword)
	auth.GET("/me", deps.AuthMW, handler.Me)
	auth.POST("/refresh", handler.Refresh)
	auth.POST("/password/change", deps.AuthMW, handler.ChangePassword)
	auth.POST("/email/confirm", handler.ConfirmEmail)
}
