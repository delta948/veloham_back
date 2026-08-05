package favorites

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewFavoriteHandler(deps.DB)
	favorites := rg.Group("/favorites", deps.AuthMW)
	favorites.GET("", h.List)
	favorites.POST("/:listingId", h.Add)
	favorites.DELETE("/:listingId", h.Delete)
}
