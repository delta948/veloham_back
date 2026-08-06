package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

type AuthHandler struct {
	db  *gorm.DB
	jwt services.JWTService
}

func NewAuthHandler(db *gorm.DB, jwt services.JWTService) AuthHandler {
	return AuthHandler{db: db, jwt: jwt}
}

type authRequest struct {
	Username string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12,max=72"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,max=72"`
}

func (h AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := models.User{Username: req.Username, Email: req.Email, PasswordHash: string(hash), Role: "user"}
	if user.Username == "" {
		user.Username = req.Email
	}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}
	token, _ := h.jwt.Generate(user.ID)
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": models.UserWithEmail(user)})
}

func (h AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if user.IsBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "user blocked"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, _ := h.jwt.Generate(user.ID)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": models.UserWithEmail(user)})
}

func (h AuthHandler) Me(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, "id = ?", middleware.CurrentUserID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, models.UserWithEmail(user))
}
