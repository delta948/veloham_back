package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

func Auth(jwtService services.JWTService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		userID, err := jwtService.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		var user models.User
		if err := db.Select("id", "is_blocked", "blocked_reason").First(&user, "id = ?", userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user is unavailable"})
			return
		}
		if user.IsBlocked {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account_blocked", "message": "Ваш аккаунт заблокирован", "reason": user.BlockedReason})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) string {
	userID, _ := c.Get("userID")
	value, _ := userID.(string)
	return value
}

func Admin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := db.First(&user, "id = ?", CurrentUserID(c)).Error; err != nil || user.Role != "admin" || user.IsBlocked {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
