package routes

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"veloham/backend/internal/common"
	"veloham/backend/internal/common/docs"
	"veloham/backend/internal/config"
	"veloham/backend/internal/handlers"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
	"veloham/backend/internal/modules/admin"
	"veloham/backend/internal/modules/auth"
	"veloham/backend/internal/modules/buy_requests"
	"veloham/backend/internal/modules/categories"
	"veloham/backend/internal/modules/chat"
	"veloham/backend/internal/modules/favorites"
	"veloham/backend/internal/modules/listings"
	"veloham/backend/internal/modules/messages"
	"veloham/backend/internal/modules/moderation"
	"veloham/backend/internal/modules/notifications"
	"veloham/backend/internal/modules/parts"
	"veloham/backend/internal/modules/reviews"
	"veloham/backend/internal/modules/search"
	"veloham/backend/internal/modules/uploads"
	"veloham/backend/internal/modules/users"
	wantedmodule "veloham/backend/internal/modules/wanted"
	"veloham/backend/internal/services"
)

func Setup(db *gorm.DB, cfg config.Config) *gin.Engine {
	_ = os.MkdirAll(cfg.UploadDir, 0755)

	r := gin.Default()
	origins := strings.Split(cfg.CORSOrigin, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	r.Static("/uploads", cfg.UploadDir)
	r.GET("/api/docs", func(c *gin.Context) {
		c.JSON(200, docs.OpenAPI())
	})
	r.GET("/api/docs/openapi.json", func(c *gin.Context) {
		c.JSON(200, docs.OpenAPI())
	})

	jwtService := services.NewJWTService(cfg.JWTSecret)
	hub := services.NewChatHub()
	authHandler := handlers.NewAuthHandler(db, jwtService)
	listingHandler := handlers.NewListingHandler(db, cfg.UploadDir)
	priceHistoryHandler := handlers.NewPriceHistoryHandler(db)
	favoriteHandler := handlers.NewFavoriteHandler(db)
	chatHandler := handlers.NewChatHandler(db, hub, jwtService)
	userHandler := handlers.NewUserHandler(db)
	reviewHandler := handlers.NewReviewHandler(db)
	adminHandler := handlers.NewReportAdminHandler(db)
	wantedHandler := handlers.NewWantedHandler(db)
	authMW := middleware.Auth(jwtService)
	adminMW := middleware.Admin(db)
	deps := common.Dependencies{
		DB: db, Config: cfg, JWT: jwtService, ChatHub: hub, AuthMW: authMW, AdminMW: adminMW,
	}

	v1 := r.Group("/api/v1")
	auth.RegisterRoutes(v1, deps)
	users.RegisterRoutes(v1, deps)
	listings.RegisterRoutes(v1, deps)
	categories.RegisterRoutes(v1, deps)
	parts.RegisterRoutes(v1, deps)
	chat.RegisterRoutes(v1, deps)
	messages.RegisterRoutes(v1, deps)
	search.RegisterRoutes(v1, deps)
	buy_requests.RegisterRoutes(v1, deps)
	wantedmodule.RegisterRoutes(v1, deps)
	favorites.RegisterRoutes(v1, deps)
	notifications.RegisterRoutes(v1, deps)
	reviews.RegisterRoutes(v1, deps)
	admin.RegisterRoutes(v1, deps)
	moderation.RegisterRoutes(v1, deps)
	uploads.RegisterRoutes(v1, deps)

	api := r.Group("/api")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.GET("/auth/me", authMW, authHandler.Me)

	api.GET("/users/:id", userHandler.Get)
	api.PUT("/users/me", authMW, userHandler.UpdateMe)
	api.GET("/users/:id/listings", userHandler.Listings)

	api.GET("/listings", listingHandler.List)
	api.GET("/listings/:id", listingHandler.Get)
	api.GET("/listings/:id/price-history", priceHistoryHandler.Get)
	api.POST("/listings", authMW, listingHandler.Create)
	api.PUT("/listings/:id", authMW, listingHandler.Update)
	api.DELETE("/listings/:id", authMW, listingHandler.Delete)
	api.PATCH("/listings/:id/status", authMW, listingHandler.PatchStatus)
	api.POST("/listings/:id/images", authMW, listingHandler.AddImages)
	api.DELETE("/images/:id", authMW, listingHandler.DeleteImage)
	api.GET("/listings/:id/matches", authMW, listingHandler.Matches)
	api.POST("/listings/:id/match-preferences", authMW, listingHandler.SaveMatchPreference)

	api.GET("/favorites", authMW, favoriteHandler.List)
	api.POST("/favorites/:listingId", authMW, favoriteHandler.Add)
	api.DELETE("/favorites/:listingId", authMW, favoriteHandler.Delete)
	api.GET("/notifications", authMW, func(c *gin.Context) {
		var rows []models.Notification
		if err := db.Where("user_id = ?", middleware.CurrentUserID(c)).Order("created_at desc").Limit(100).Find(&rows).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, rows)
	})
	api.PATCH("/notifications/:id/read", authMW, func(c *gin.Context) {
		db.Model(&models.Notification{}).Where("id = ? AND user_id = ?", c.Param("id"), middleware.CurrentUserID(c)).Update("is_read", true)
		c.Status(204)
	})

	api.GET("/chats", authMW, chatHandler.List)
	api.GET("/chats/:id", authMW, chatHandler.Get)
	api.GET("/chats/:id/messages", authMW, chatHandler.Messages)
	api.POST("/chats", authMW, chatHandler.Create)

	api.POST("/reviews", authMW, reviewHandler.Create)
	api.GET("/users/:id/reviews", reviewHandler.ListByUser)
	api.POST("/reports", authMW, adminHandler.CreateReport)

	api.GET("/wanted", wantedHandler.List)
	api.POST("/wanted", authMW, wantedHandler.Create)
	api.GET("/wanted/:id", wantedHandler.Get)
	api.POST("/wanted/:id/offers", authMW, wantedHandler.Offer)
	api.PATCH("/wanted/:id/close", authMW, wantedHandler.Close)
	api.GET("/want-to-buy", wantedHandler.List)
	api.POST("/want-to-buy", authMW, wantedHandler.Create)
	api.GET("/want-to-buy/:id", wantedHandler.Get)

	api.GET("/admin/reports", authMW, adminMW, adminHandler.Reports)
	api.GET("/admin/users", authMW, adminMW, adminHandler.Users)
	api.PATCH("/admin/users/:id/block", authMW, adminMW, adminHandler.BlockUser)
	api.GET("/admin/listings", authMW, adminMW, adminHandler.Listings)
	api.GET("/admin/price-history", authMW, adminMW, priceHistoryHandler.AdminList)
	api.DELETE("/admin/listings/:id", authMW, adminMW, adminHandler.DeleteListing)

	r.GET("/ws/chats/:id", chatHandler.WebSocket)
	return r
}
