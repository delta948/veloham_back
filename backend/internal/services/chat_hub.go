package services

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ChatHub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]bool
}

func NewChatHub() *ChatHub {
	return &ChatHub{rooms: make(map[string]map[*websocket.Conn]bool)}
}

func (h *ChatHub) Join(chatID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[chatID] == nil {
		h.rooms[chatID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[chatID][conn] = true
}

func (h *ChatHub) Leave(chatID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[chatID], conn)
	_ = conn.Close()
}

func (h *ChatHub) Broadcast(chatID string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.rooms[chatID] {
		_ = conn.WriteJSON(payload)
	}
}
