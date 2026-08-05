package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

type PriceHistoryHandler struct {
	db      *gorm.DB
	service services.PriceHistoryService
}

func NewPriceHistoryHandler(db *gorm.DB) PriceHistoryHandler {
	return PriceHistoryHandler{db, services.PriceHistoryService{DB: db}}
}

func (h PriceHistoryHandler) Get(c *gin.Context) {
	result, err := h.service.Get(c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h PriceHistoryHandler) AdminList(c *gin.Context) {
	var rows []models.ListingPriceHistory
	q := h.db.Preload("Listing.User").Preload("User").Order("changed_at desc")
	if c.Query("suspicious") == "true" {
		q = q.Where("suspicious = true")
	}
	if err := q.Limit(500).Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	type adminRow struct {
		models.ListingPriceHistory
		ChangeCount int64 `json:"change_count"`
	}
	result := make([]adminRow, 0, len(rows))
	for _, row := range rows {
		var count int64
		h.db.Model(&models.ListingPriceHistory{}).Where("listing_id = ?", row.ListingID).Count(&count)
		result = append(result, adminRow{row, count})
	}
	c.JSON(http.StatusOK, result)
}

func actorIsAdmin(db *gorm.DB, userID string) bool {
	var user models.User
	return db.Select("role").First(&user, "id = ?", userID).Error == nil && user.Role == "admin"
}

func currentActor(c *gin.Context) string { return middleware.CurrentUserID(c) }
