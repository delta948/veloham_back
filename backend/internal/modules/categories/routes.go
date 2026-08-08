package categories

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, _ common.Dependencies) {
	handler := NewHandler(NewService())

	categories := rg.Group("/categories")
	categories.GET("", handler.List)
}
