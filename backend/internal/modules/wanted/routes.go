package wanted

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	handler := NewHandler(NewService(deps.DB))

	wanted := rg.Group("/wanted")
	wanted.GET("", handler.List)
	wanted.POST("", deps.AuthMW, handler.Create)
	wanted.GET("/:id", handler.Get)
	wanted.POST("/:id/offers", deps.AuthMW, handler.Offer)
	wanted.PATCH("/:id/close", deps.AuthMW, handler.Close)

	wantToBuy := rg.Group("/want-to-buy")
	wantToBuy.GET("", handler.List)
	wantToBuy.POST("", deps.AuthMW, handler.Create)
	wantToBuy.GET("/:id", handler.Get)
}
