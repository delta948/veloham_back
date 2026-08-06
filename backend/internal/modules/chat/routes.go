package chat

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewChatHandler(deps.DB, deps.ChatHub, deps.JWT, deps.Config.CORSOrigin)
	chats := rg.Group("/chats", deps.AuthMW)
	chats.GET("", h.List)
	chats.GET("/:id", h.Get)
	chats.POST("", h.Create)
	chats.GET("/:id/messages", h.Messages)
}
