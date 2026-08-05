package notifications

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	notifications := rg.Group("/notifications", deps.AuthMW)
	notifications.GET("", func(c *gin.Context) {
		c.JSON(200, []any{})
	})
	notifications.PATCH("/:id/read", func(c *gin.Context) {
		c.Status(204)
	})
}
