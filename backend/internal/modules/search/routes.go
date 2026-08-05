package search

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewListingHandler(deps.DB, deps.Config.UploadDir)
	search := rg.Group("/search")
	search.GET("/listings", h.List)
	search.GET("", h.List)
}
