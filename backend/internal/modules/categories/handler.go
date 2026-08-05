package categories

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.List())
}

func (h *Handler) Create(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"status": "category management boundary is ready"})
}
