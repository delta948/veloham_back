package handlers

import (
	"net/http"
	"strings"

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
	origins  map[string]struct{}
}

func NewChatHandler(db *gorm.DB, hub *services.ChatHub, jwt services.JWTService, allowedOrigins string) ChatHandler {
	originList := strings.Split(allowedOrigins, ",")
	handler := ChatHandler{
		db:      db,
		hub:     hub,
		jwt:     jwt,
		origins: make(map[string]struct{}, len(originList)),
	}
	for _, origin := range originList {
		if origin = strings.TrimSpace(origin); origin != "" {
			handler.origins[origin] = struct{}{}
		}
	}
	handler.upgrader = websocket.Upgrader{CheckOrigin: handler.checkOrigin}
	return handler
}

func (h ChatHandler) checkOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	_, allowed := h.origins[origin]
	return allowed
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
	token := websocketToken(c.Request)
	userID, err := h.jwt.Parse(token)
	chatID := c.Param("id")
	if err != nil || !h.canAccessChat(chatID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "chat access denied"})
		return
	}
	responseHeaders := http.Header{}
	if len(websocket.Subprotocols(c.Request)) > 0 {
		responseHeaders.Set("Sec-WebSocket-Protocol", "access_token")
	}
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, responseHeaders)
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

func websocketToken(r *http.Request) string {
	protocols := websocket.Subprotocols(r)
	if len(protocols) == 2 && protocols[0] == "access_token" {
		return protocols[1]
	}
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}

func (h ChatHandler) canAccessChat(chatID, userID string) bool {
	var count int64
	h.db.Model(&models.Chat{}).Where("id = ? AND (buyer_id = ? OR seller_id = ?)", chatID, userID, userID).Count(&count)
	return count > 0
}
