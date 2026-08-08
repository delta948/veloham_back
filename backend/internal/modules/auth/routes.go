package auth

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	handler := NewHandler(NewService(deps.DB, deps.JWT, deps.Config))

	auth := rg.Group("/auth")
	auth.POST("/register", deps.AuthRateMW, handler.Register)
	auth.POST("/register/verify", deps.AuthRateMW, handler.VerifyRegistration)
	auth.POST("/register/resend", deps.AuthRateMW, handler.ResendRegistration)
	auth.POST("/login", deps.AuthRateMW, handler.Login)
	auth.GET("/me", deps.AuthMW, handler.Me)
	auth.POST("/password/change", deps.AuthMW, handler.ChangePassword)
}
