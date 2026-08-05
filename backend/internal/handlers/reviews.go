package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type ReviewHandler struct {
	db *gorm.DB
}

func NewReviewHandler(db *gorm.DB) ReviewHandler {
	return ReviewHandler{db: db}
}

func (h ReviewHandler) Create(c *gin.Context) {
	var req models.Review
	if err := c.ShouldBindJSON(&req); err != nil || req.SellerID == "" || req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "seller_id and rating 1-5 are required"})
		return
	}
	req.AuthorID = middleware.CurrentUserID(c)
	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.recalculateSellerRating(req.SellerID)
	c.JSON(http.StatusCreated, req)
}

func (h ReviewHandler) ListByUser(c *gin.Context) {
	var reviews []models.Review
	if err := h.db.Preload("Author").Where("seller_id = ?", c.Param("id")).Order("created_at desc").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

func (h ReviewHandler) recalculateSellerRating(sellerID string) {
	var avg float64
	h.db.Model(&models.Review{}).Select("COALESCE(AVG(rating), 0)").Where("seller_id = ?", sellerID).Scan(&avg)
	h.db.Model(&models.User{}).Where("id = ?", sellerID).Update("rating", avg)
}
