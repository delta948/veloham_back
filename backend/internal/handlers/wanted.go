package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type WantedHandler struct {
	db *gorm.DB
}

func NewWantedHandler(db *gorm.DB) WantedHandler {
	return WantedHandler{db: db}
}

func (h WantedHandler) List(c *gin.Context) {
	var requests []models.WantedRequest
	q := h.db.Preload("User").Order("created_at desc")
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if category := c.Query("category"); category != "" {
		q = q.Where("category = ?", category)
	}
	if err := q.Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range requests {
		hydrateWantedContract(&requests[i])
	}
	c.JSON(http.StatusOK, requests)
}

func (h WantedHandler) Create(c *gin.Context) {
	var req models.WantedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeWantedContract(&req)
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Category) == "" || strings.TrimSpace(req.City) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, category and city are required"})
		return
	}
	if req.MinBudget > 0 && req.MaxBudget > 0 && req.MinBudget > req.MaxBudget {
		c.JSON(http.StatusBadRequest, gin.H{"error": "budget_min must be less than budget_max"})
		return
	}
	if req.RiderHeight < 0 || req.Height < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "height must be positive"})
		return
	}
	req.UserID = middleware.CurrentUserID(c)
	if req.Status == "" {
		req.Status = "active"
	}
	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hydrateWantedContract(&req)
	c.JSON(http.StatusCreated, req)
}

func (h WantedHandler) Get(c *gin.Context) {
	var req models.WantedRequest
	if err := h.db.Preload("User").Preload("Offers.Seller").Preload("Offers.Listing.Images").First(&req, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wanted request not found"})
		return
	}
	hydrateWantedContract(&req)
	var matches []models.Listing
	q := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").
		Where("status = ? AND category = ?", "active", req.Category)
	if req.MinBudget > 0 {
		q = q.Where("price >= ?", req.MinBudget)
	}
	if req.MaxBudget > 0 {
		q = q.Where("price <= ?", req.MaxBudget)
	}
	if req.City != "" {
		q = q.Where("city = ?", req.City)
	}
	if req.FrameSize != "" {
		q = q.Where("frame_size ILIKE ? OR build_cards.size ILIKE ?", req.FrameSize, req.FrameSize)
		q = q.Joins("LEFT JOIN build_cards ON build_cards.listing_id = listings.id")
	}
	if req.RiderHeight > 0 {
		q = q.Where("(rider_height_min = 0 OR rider_height_min <= ?) AND (rider_height_max = 0 OR rider_height_max >= ?)", req.RiderHeight, req.RiderHeight)
	}
	if req.PreferredBikeType != "" {
		q = q.Where("bike_type = ?", req.PreferredBikeType)
	}
	q.Order("created_at desc").Find(&matches)
	for i := range matches {
		hydrateListingContract(&matches[i])
	}
	c.JSON(http.StatusOK, gin.H{"request": req, "matches": matches})
}

func (h WantedHandler) Offer(c *gin.Context) {
	var req struct {
		ListingID string `json:"listing_id"`
		Message   string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ListingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listing_id is required"})
		return
	}
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", req.ListingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only listing owner can offer this item"})
		return
	}
	offer := models.WantedOffer{WantedID: c.Param("id"), SellerID: listing.UserID, ListingID: listing.ID, Message: req.Message}
	if err := h.db.Create(&offer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("Seller").Preload("Listing.Images").First(&offer, "id = ?", offer.ID)
	c.JSON(http.StatusCreated, offer)
}

func (h WantedHandler) Close(c *gin.Context) {
	var req models.WantedRequest
	if err := h.db.First(&req, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wanted request not found"})
		return
	}
	if req.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can close request"})
		return
	}
	h.db.Model(&req).Update("status", "closed")
	hydrateWantedContract(&req)
	c.JSON(http.StatusOK, req)
}

func normalizeWantedContract(req *models.WantedRequest) {
	if req.BudgetMin == 0 {
		req.BudgetMin = req.MinBudget
	}
	if req.BudgetMax == 0 {
		req.BudgetMax = req.MaxBudget
	}
	if req.MinBudget == 0 {
		req.MinBudget = req.BudgetMin
	}
	if req.MaxBudget == 0 {
		req.MaxBudget = req.BudgetMax
	}
	if req.Height == 0 {
		req.Height = req.RiderHeight
	}
	if req.RiderHeight == 0 {
		req.RiderHeight = req.Height
	}
	req.PreferredBikeType = normalizeBikeType(req.PreferredBikeType)
}

func hydrateWantedContract(req *models.WantedRequest) {
	normalizeWantedContract(req)
}
