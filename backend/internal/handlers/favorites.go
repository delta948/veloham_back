package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type FavoriteHandler struct {
	db *gorm.DB
}

func NewFavoriteHandler(db *gorm.DB) FavoriteHandler {
	return FavoriteHandler{db: db}
}

func (h FavoriteHandler) List(c *gin.Context) {
	var favorites []models.Favorite
	if err := h.db.Preload("Listing.Images").Preload("Listing.User").Where("user_id = ?", middleware.CurrentUserID(c)).Order("created_at desc").Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, favorites)
}

func (h FavoriteHandler) Add(c *gin.Context) {
	favorite := models.Favorite{UserID: middleware.CurrentUserID(c), ListingID: c.Param("listingId")}
	if err := h.db.FirstOrCreate(&favorite, favorite).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, favorite)
}

func (h FavoriteHandler) Delete(c *gin.Context) {
	h.db.Where("user_id = ? AND listing_id = ?", middleware.CurrentUserID(c), c.Param("listingId")).Delete(&models.Favorite{})
	c.Status(http.StatusNoContent)
}
