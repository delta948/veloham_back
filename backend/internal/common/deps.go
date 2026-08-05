package common

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/config"
	"veloham/backend/internal/services"
)

type Dependencies struct {
	DB      *gorm.DB
	Config  config.Config
	JWT     services.JWTService
	ChatHub *services.ChatHub
	AuthMW  gin.HandlerFunc
	AdminMW gin.HandlerFunc
}
