package categories

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	handler := NewHandler(NewService())

	categories := rg.Group("/categories")
	categories.GET("", handler.List)
	categories.POST("", deps.AuthMW, deps.AdminMW, handler.Create)
}
