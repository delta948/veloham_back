package listings

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewListingHandler(deps.DB, deps.Config.UploadDir)
	priceHistory := handlers.NewPriceHistoryHandler(deps.DB)
	listings := rg.Group("/listings")
	listings.GET("", h.List)
	listings.GET("/:id", h.Get)
	listings.GET("/:id/price-history", priceHistory.Get)
	listings.POST("", deps.AuthMW, h.Create)
	listings.PUT("/:id", deps.AuthMW, h.Update)
	listings.DELETE("/:id", deps.AuthMW, h.Delete)
	listings.PATCH("/:id/status", deps.AuthMW, h.PatchStatus)
	listings.PATCH("/:id/archive", deps.AuthMW, h.Archive)
	listings.PATCH("/:id/bump", deps.AuthMW, h.Bump)
	listings.POST("/:id/images", deps.AuthMW, h.AddImages)
	listings.DELETE("/images/:id", deps.AuthMW, h.DeleteImage)
	listings.GET("/:id/matches", deps.AuthMW, h.Matches)
	listings.POST("/:id/match-preferences", deps.AuthMW, h.SaveMatchPreference)
}
