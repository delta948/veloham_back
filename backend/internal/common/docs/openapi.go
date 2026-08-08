package docs

import "github.com/gin-gonic/gin"

func OpenAPI() gin.H {
	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":   "VELOHAM API",
			"version": "1.0.0",
		},
		"paths": gin.H{
			"/api/v1/auth/register":               gin.H{"post": gin.H{"summary": "Register user"}},
			"/api/v1/auth/login":                  gin.H{"post": gin.H{"summary": "Login user"}},
			"/api/v1/auth/me":                     gin.H{"get": gin.H{"summary": "Current user"}},
			"/api/v1/users/profile":               gin.H{"get": gin.H{"summary": "Current user profile"}},
			"/api/v1/listings":                    gin.H{"get": gin.H{"summary": "List listings"}, "post": gin.H{"summary": "Create listing"}},
			"/api/v1/listings/{id}":               gin.H{"get": gin.H{"summary": "Get listing"}, "put": gin.H{"summary": "Update own listing"}, "delete": gin.H{"summary": "Delete own listing"}},
			"/api/v1/listings/{id}/price-history": gin.H{"get": gin.H{"summary": "Get listing price history"}},
			"/api/v1/categories":                  gin.H{"get": gin.H{"summary": "List categories"}},
			"/api/v1/uploads":                     gin.H{"post": gin.H{"summary": "Upload files"}},
			"/api/v1/favorites":                   gin.H{"get": gin.H{"summary": "List favorites"}},
			"/api/v1/chats":                       gin.H{"get": gin.H{"summary": "List conversations"}, "post": gin.H{"summary": "Create conversation"}},
			"/api/v1/notifications":               gin.H{"get": gin.H{"summary": "List current user notifications"}},
			"/api/v1/notifications/{id}/read":     gin.H{"patch": gin.H{"summary": "Mark own notification as read"}},
			"/api/v1/wanted":                      gin.H{"get": gin.H{"summary": "List wanted posts"}, "post": gin.H{"summary": "Create wanted post"}},
			"/api/v1/moderation":                  gin.H{"get": gin.H{"summary": "Moderation boundary"}},
		},
	}
}
