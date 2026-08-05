package parts

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewListingHandler(deps.DB, deps.Config.UploadDir)
	parts := rg.Group("/parts")
	parts.GET("", h.List)
	parts.GET("/:id", h.Get)
	parts.POST("", deps.AuthMW, h.Create)
	parts.PUT("/:id", deps.AuthMW, h.Update)
	parts.DELETE("/:id", deps.AuthMW, h.Delete)
}
