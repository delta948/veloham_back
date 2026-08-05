package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

type ChatHandler struct {
	db       *gorm.DB
	hub      *services.ChatHub
	jwt      services.JWTService
	upgrader websocket.Upgrader
}

func NewChatHandler(db *gorm.DB, hub *services.ChatHub, jwt services.JWTService) ChatHandler {
	return ChatHandler{
		db:  db,
		hub: hub,
		jwt: jwt,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type createChatRequest struct {
	ListingID string `json:"listing_id" binding:"required"`
}

func (h ChatHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	var chats []models.Chat
	err := h.db.Preload("Buyer").Preload("Seller").Preload("Listing.Images").
		Where("buyer_id = ? OR seller_id = ?", userID, userID).
		Order("created_at desc").Find(&chats).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}

func (h ChatHandler) Get(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	var chat models.Chat
	err := h.db.Preload("Buyer").Preload("Seller").Preload("Listing.Images").
		Where("id = ? AND (buyer_id = ? OR seller_id = ?)", c.Param("id"), userID, userID).
		First(&chat).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
		return
	}
	c.JSON(http.StatusOK, chat)
}

func (h ChatHandler) Create(c *gin.Context) {
	var req createChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", req.ListingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	buyerID := middleware.CurrentUserID(c)
	if buyerID == listing.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "seller cannot chat with self"})
		return
	}
	chat := models.Chat{BuyerID: buyerID, SellerID: listing.UserID, ListingID: listing.ID}
	if err := h.db.Where(chat).FirstOrCreate(&chat).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("Buyer").Preload("Seller").Preload("Listing.Images").First(&chat, "id = ?", chat.ID)
	c.JSON(http.StatusCreated, chat)
}

func (h ChatHandler) Messages(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	if !h.canAccessChat(c.Param("id"), userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "chat access denied"})
		return
	}
	var messages []models.Message
	if err := h.db.Preload("Sender").Where("chat_id = ?", c.Param("id")).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

type incomingMessage struct {
	Text string `json:"text"`
}

func (h ChatHandler) WebSocket(c *gin.Context) {
	userID, err := h.jwt.Parse(c.Query("token"))
	chatID := c.Param("id")
	if err != nil || !h.canAccessChat(chatID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "chat access denied"})
		return
	}
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.Join(chatID, conn)
	defer h.hub.Leave(chatID, conn)

	for {
		var input incomingMessage
		if err := conn.ReadJSON(&input); err != nil {
			return
		}
		if input.Text == "" {
			continue
		}
		msg := models.Message{ChatID: chatID, SenderID: userID, Text: input.Text}
		if err := h.db.Create(&msg).Error; err != nil {
			continue
		}
		h.db.Preload("Sender").First(&msg, "id = ?", msg.ID)
		h.hub.Broadcast(chatID, msg)
	}
}

func (h ChatHandler) canAccessChat(chatID, userID string) bool {
	var count int64
	h.db.Model(&models.Chat{}).Where("id = ? AND (buyer_id = ? OR seller_id = ?)", chatID, userID, userID).Count(&count)
	return count > 0
}
