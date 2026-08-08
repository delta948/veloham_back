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
		if errors.Is(err, ErrResendCooldown) || errors.Is(err, ErrVerificationLimit) {
			response.Error(c, http.StatusTooManyRequests, err.Error())
			return
		}
		if errors.Is(err, errEmailNotConfigured) {
			response.Error(c, http.StatusServiceUnavailable, "email delivery is not configured")
			return
		}
		response.Error(c, http.StatusBadGateway, "failed to send verification email")
		return
	}
	c.JSON(http.StatusAccepted, res)
}

func (h *Handler) VerifyRegistration(c *gin.Context) {
	var req VerifyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.service.VerifyRegistration(c.Request.Context(), req.VerificationID, req.Code)
	if err != nil {
		if errors.Is(err, ErrVerificationLimit) {
			response.Error(c, http.StatusTooManyRequests, err.Error())
			return
		}
		if errors.Is(err, ErrVerification) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusBadGateway, "failed to verify email code")
		return
	}
	response.Created(c, res)
}

func (h *Handler) ResendRegistration(c *gin.Context) {
	var req ResendRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.service.ResendRegistration(c.Request.Context(), req.VerificationID)
	if err != nil {
		if errors.Is(err, ErrResendCooldown) || errors.Is(err, ErrVerificationLimit) {
			response.Error(c, http.StatusTooManyRequests, err.Error())
			return
		}
		if errors.Is(err, ErrVerification) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, errEmailNotConfigured) {
			response.Error(c, http.StatusServiceUnavailable, "email delivery is not configured")
			return
		}
		response.Error(c, http.StatusBadGateway, "failed to resend verification email")
		return
	}
	c.JSON(http.StatusAccepted, res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Identifier() == "" {
		response.Error(c, http.StatusBadRequest, "login is required")
		return
	}
	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserBlocked) {
			var blocked UserBlockedError
			errors.As(err, &blocked)
			c.JSON(http.StatusForbidden, gin.H{"error": "account_blocked", "message": "Ваш аккаунт заблокирован", "reason": blocked.Reason})
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
