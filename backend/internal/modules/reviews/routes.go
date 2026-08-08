package reviews

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewReviewHandler(deps.DB)
	reviews := rg.Group("/reviews")
	reviews.POST("", deps.AuthMW, h.Create)
	reviews.GET("/users/:id", h.ListByUser)
	rg.GET("/users/:id/reviews", h.ListByUser)
}
