package moderation

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	handler := NewHandler(NewService(deps.DB))

	moderation := rg.Group("/moderation", deps.AuthMW, deps.AdminMW)
	moderation.GET("/reports", handler.Reports)
	moderation.GET("/listings", handler.Listings)
	moderation.POST("/listings/:id/hide", handler.HideListing)
	moderation.DELETE("/listings/:id", handler.DeleteListing)
}
