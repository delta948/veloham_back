package notifications

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	handler := NewHandler(NewService(NewRepository(deps.DB)))
	notifications := rg.Group("/notifications", deps.AuthMW)
	notifications.GET("", handler.List)
	notifications.PATCH("/read-all", handler.MarkAllRead)
	notifications.PATCH("/:id/read", handler.MarkRead)
}
