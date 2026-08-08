package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type ReportAdminHandler struct {
	db *gorm.DB
}

func NewReportAdminHandler(db *gorm.DB) ReportAdminHandler {
	return ReportAdminHandler{db: db}
}

func (h ReportAdminHandler) CreateReport(c *gin.Context) {
	var req models.Report
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}
	req.ReporterID = middleware.CurrentUserID(c)
	req.Status = "new"
	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h ReportAdminHandler) Reports(c *gin.Context) {
	var reports []models.Report
	if err := h.db.Preload("Reporter").Preload("Listing.Images").Order("created_at desc").Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h ReportAdminHandler) Users(c *gin.Context) {
	var users []models.User
	query := h.db.Order("created_at desc").Limit(100)
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR phone LIKE ?", pattern, pattern, "%"+search+"%")
	}
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	privateUsers := make([]models.PrivateUser, 0, len(users))
	for _, user := range users {
		privateUsers = append(privateUsers, models.UserWithEmail(user))
	}
	c.JSON(http.StatusOK, privateUsers)
}

func (h ReportAdminHandler) BlockUser(c *gin.Context) {
	var req struct {
		IsBlocked bool   `json:"is_blocked"`
		Reason    string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if req.IsBlocked && reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "block reason is required"})
		return
	}
	if len([]rune(reason)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "block reason is too long"})
		return
	}
	if c.Param("id") == middleware.CurrentUserID(c) && req.IsBlocked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "administrator cannot block own account"})
		return
	}
	if !req.IsBlocked {
		reason = ""
	}
	var found bool
	err := h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).Where("id = ?", c.Param("id")).Updates(map[string]any{"is_blocked": req.IsBlocked, "blocked_reason": reason})
		if result.Error != nil {
			return result.Error
		}
		found = result.RowsAffected > 0
		if !found {
			return nil
		}
		action := "unblocked"
		if req.IsBlocked {
			action = "blocked"
		}
		return tx.Create(&models.UserBlockEvent{UserID: c.Param("id"), AdminID: middleware.CurrentUserID(c), Action: action, Reason: reason}).Error
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h ReportAdminHandler) BlockEvents(c *gin.Context) {
	var events []models.UserBlockEvent
	if err := h.db.Preload("User").Preload("Admin").Order("created_at desc").Limit(200).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h ReportAdminHandler) Listings(c *gin.Context) {
	var listings []models.Listing
	if err := h.db.Preload("Images").Preload("User").Order("created_at desc").Find(&listings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listings)
}

func (h ReportAdminHandler) DeleteListing(c *gin.Context) {
	if err := deleteListingCascade(h.db, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h ReportAdminHandler) Stats(c *gin.Context) {
	var usersCount, listingsCount, reportsCount, blockedCount, pendingPayments, paidPayments int64
	h.db.Model(&models.User{}).Count(&usersCount)
	h.db.Model(&models.Listing{}).Count(&listingsCount)
	h.db.Model(&models.Report{}).Count(&reportsCount)
	h.db.Model(&models.User{}).Where("is_blocked = true").Count(&blockedCount)
	h.db.Model(&models.ListingPlacement{}).Where("status = 'pending'").Count(&pendingPayments)
	h.db.Model(&models.ListingPlacement{}).Where("status = 'paid'").Count(&paidPayments)
	c.JSON(http.StatusOK, gin.H{
		"users":            usersCount,
		"listings":         listingsCount,
		"reports":          reportsCount,
		"blocked_users":    blockedCount,
		"pending_payments": pendingPayments,
		"paid_payments":    paidPayments,
	})
}

func (h ReportAdminHandler) HideListing(c *gin.Context) {
	if err := h.db.Model(&models.Listing{}).Where("id = ?", c.Param("id")).Update("status", "hidden").Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
