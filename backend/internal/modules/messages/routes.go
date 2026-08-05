package messages

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewChatHandler(deps.DB, deps.ChatHub, deps.JWT)
	rg.GET("/messages/chats/:id", deps.AuthMW, h.Messages)
}
