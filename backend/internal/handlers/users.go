package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type UserHandler struct {
	db        *gorm.DB
	uploadDir string
}

func NewUserHandler(db *gorm.DB, uploadDir string) UserHandler {
	return UserHandler{db: db, uploadDir: uploadDir}
}

func (h UserHandler) UploadAvatar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageBytes+(1<<20))
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar image is required"})
		return
	}
	if file.Size > maxImageBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": errImageTooLarge.Error()})
		return
	}
	extension, err := imageExtension(file)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	suffix, err := randomHex(12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create avatar name"})
		return
	}
	userID := middleware.CurrentUserID(c)
	filename := fmt.Sprintf("avatar-%s-%s%s", userID, suffix, extension)
	destination := filepath.Join(h.uploadDir, filename)
	if err := saveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save avatar"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		_ = os.Remove(destination)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	previous := user.AvatarURL
	avatarURL := "/uploads/" + filename
	if err := h.db.Model(&user).Update("avatar_url", avatarURL).Error; err != nil {
		_ = os.Remove(destination)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update avatar"})
		return
	}
	if strings.HasPrefix(previous, "/uploads/avatar-") {
		_ = os.Remove(filepath.Join(h.uploadDir, filepath.Base(previous)))
	}
	user.AvatarURL = avatarURL
	c.JSON(http.StatusOK, models.UserWithEmail(user))
}

func (h UserHandler) Profile(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, "id = ?", middleware.CurrentUserID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, models.UserWithEmail(user))
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
		Username string `json:"username"`
		City     string `json:"city"`
		Contact  string `json:"contact"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{
		"username": req.Username, "city": req.City, "contact": req.Contact,
	}
	if req.Password != "" {
		if len(req.Password) < 6 || len(req.Password) > 72 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must contain 6 to 72 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
			return
		}
		updates["password_hash"] = string(hash)
	}
	var user models.User
	if err := h.db.First(&user, "id = ?", middleware.CurrentUserID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	h.db.Model(&user).Updates(updates)
	h.db.First(&user, "id = ?", user.ID)
	c.JSON(http.StatusOK, models.UserWithEmail(user))
}

func (h UserHandler) Listings(c *gin.Context) {
	var listings []models.Listing
	if err := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").
		Where("user_id = ? AND status NOT IN ?", c.Param("id"), []string{"hidden", "pending_payment"}).
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

func (h UserHandler) MySales(c *gin.Context) {
	var listings []models.Listing
	if err := h.db.Preload("Images").Where("user_id = ? AND status = ?", middleware.CurrentUserID(c), "sold").Find(&listings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listings)
}
