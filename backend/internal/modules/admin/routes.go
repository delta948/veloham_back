package admin

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewReportAdminHandler(deps.DB)
	admin := rg.Group("/admin", deps.AuthMW, deps.AdminMW)
	admin.GET("/reports", h.Reports)
	admin.GET("/users", h.Users)
	admin.PATCH("/users/:id/block", h.BlockUser)
	admin.GET("/listings", h.Listings)
	admin.DELETE("/listings/:id", h.DeleteListing)
	admin.GET("/stats", h.Stats)
	admin.POST("/moderation/listings/:id/hide", h.HideListing)
}
