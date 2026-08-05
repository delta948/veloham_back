package handlers

import (
	"net/http"

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
	if err := h.db.Order("created_at desc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h ReportAdminHandler) BlockUser(c *gin.Context) {
	var req struct {
		IsBlocked bool `json:"is_blocked"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.db.Model(&models.User{}).Where("id = ?", c.Param("id")).Update("is_blocked", req.IsBlocked).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
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
	var usersCount, listingsCount, reportsCount int64
	h.db.Model(&models.User{}).Count(&usersCount)
	h.db.Model(&models.Listing{}).Count(&listingsCount)
	h.db.Model(&models.Report{}).Count(&reportsCount)
	c.JSON(http.StatusOK, gin.H{
		"users":    usersCount,
		"listings": listingsCount,
		"reports":  reportsCount,
	})
}

func (h ReportAdminHandler) HideListing(c *gin.Context) {
	if err := h.db.Model(&models.Listing{}).Where("id = ?", c.Param("id")).Update("status", "hidden").Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
