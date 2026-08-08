package notifications

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"veloham/backend/internal/middleware"
)

type Handler struct{ service Service }

func NewHandler(service Service) Handler { return Handler{service: service} }

func (h Handler) List(c *gin.Context) {
	rows, err := h.service.List(middleware.CurrentUserID(c))
	if err != nil {
		log.Printf("list notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load notifications"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h Handler) MarkRead(c *gin.Context) {
	found, err := h.service.MarkRead(middleware.CurrentUserID(c), c.Param("id"))
	if err != nil {
		log.Printf("mark notification read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h Handler) MarkAllRead(c *gin.Context) {
	if err := h.service.MarkAllRead(middleware.CurrentUserID(c)); err != nil {
		log.Printf("mark all notifications read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notifications"})
		return
	}
	c.Status(http.StatusNoContent)
}
