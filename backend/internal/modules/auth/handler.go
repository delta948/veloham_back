package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common/response"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "registration failed")
		return
	}
	response.Created(c, res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserBlocked) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusUnauthorized, ErrInvalidCredentials.Error())
		return
	}
	response.OK(c, res)
}

func (h *Handler) Me(c *gin.Context) {
	user, err := h.service.CurrentUser(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		response.Error(c, http.StatusNotFound, ErrUserNotFound.Error())
		return
	}
	response.OK(c, models.UserWithEmail(user))
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req PasswordForgotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.AcceptPasswordRecovery(c.Request.Context(), req); err != nil {
		response.Error(c, http.StatusInternalServerError, "password recovery failed")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "password recovery request accepted"})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), middleware.CurrentUserID(c), req); err != nil {
		response.Error(c, http.StatusInternalServerError, "password change failed")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "password changed"})
}

func (h *Handler) ConfirmEmail(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"status": "email confirmation endpoint reserved"})
}

func (h *Handler) Refresh(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "refresh token service is planned"})
}
