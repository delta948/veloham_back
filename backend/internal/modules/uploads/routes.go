package uploads

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewListingHandler(deps.DB, deps.Config.UploadDir)
	uploads := rg.Group("/uploads", deps.AuthMW)
	uploads.POST("/listings/:id/images", h.AddImages)
	uploads.DELETE("/images/:id", h.DeleteImage)
}
