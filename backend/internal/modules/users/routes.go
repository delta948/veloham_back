package users

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
	"veloham/backend/internal/handlers"
)

type Repository interface{}
type Service interface{}

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := handlers.NewUserHandler(deps.DB, deps.Config.UploadDir)
	users := rg.Group("/users")
	users.GET("/profile", deps.AuthMW, h.Profile)
	users.PUT("/me", deps.AuthMW, h.UpdateMe)
	users.POST("/me/avatar", deps.AuthMW, h.UploadAvatar)
	users.GET("/me/listings", deps.AuthMW, h.MyListings)
	users.GET("/me/sales", deps.AuthMW, h.MySales)
	users.PUT("/me/settings", deps.AuthMW, h.UpdateMe)
	users.GET("/:id/listings", h.Listings)
	users.GET("/:id", h.Get)
}
