package buy_requests

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewWantedHandler(deps.DB)
	wanted := rg.Group("/buy-requests")
	wanted.GET("", h.List)
	wanted.POST("", deps.AuthMW, h.Create)
	wanted.GET("/:id", h.Get)
	wanted.POST("/:id/offers", deps.AuthMW, h.Offer)
	wanted.PATCH("/:id/close", deps.AuthMW, h.Close)
}
