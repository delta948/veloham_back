package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) UserHandler {
	return UserHandler{db: db}
}

func (h UserHandler) Profile(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, "id = ?", middleware.CurrentUserID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h UserHandler) Get(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h UserHandler) UpdateMe(c *gin.Context) {
	var req struct {
		Username  string `json:"username"`
		City      string `json:"city"`
		Contact   string `json:"contact"`
		AvatarURL string `json:"avatar_url"`
		Password  string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{
		"username": req.Username, "city": req.City, "contact": req.Contact, "avatar_url": req.AvatarURL,
	}
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		updates["password_hash"] = string(hash)
	}
	var user models.User
	if err := h.db.First(&user, "id = ?", middleware.CurrentUserID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	h.db.Model(&user).Updates(updates)
	h.db.First(&user, "id = ?", user.ID)
	c.JSON(http.StatusOK, user)
}

func (h UserHandler) Listings(c *gin.Context) {
	var listings []models.Listing
	if err := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").
		Where("user_id = ? AND status <> ?", c.Param("id"), "hidden").
		Order("created_at desc").Find(&listings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listings)
}

func (h UserHandler) MyListings(c *gin.Context) {
	var listings []models.Listing
	if err := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").
		Where("user_id = ?", middleware.CurrentUserID(c)).
		Order("created_at desc").Find(&listings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listings)
}

func (h UserHandler) MyHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"purchases": []any{}, "sales": []any{}, "note": "deal history module boundary is ready"})
}

func (h UserHandler) MyPurchases(c *gin.Context) {
	c.JSON(http.StatusOK, []any{})
}

func (h UserHandler) MySales(c *gin.Context) {
	var listings []models.Listing
	if err := h.db.Preload("Images").Where("user_id = ? AND status = ?", middleware.CurrentUserID(c), "sold").Find(&listings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listings)
}
