package auth

import (
	"strings"

	"veloham/backend/internal/models"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=80"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	City     string `json:"city"`
	Contact  string `json:"contact"`
}

type RegistrationStartResponse struct {
	VerificationID string `json:"verification_id"`
	Email          string `json:"email"`
	ExpiresIn      int    `json:"expires_in"`
	DevCode        string `json:"dev_code,omitempty"`
}

type VerifyRegistrationRequest struct {
	VerificationID string `json:"verification_id" binding:"required,uuid"`
	Code           string `json:"code" binding:"required,min=4,max=12"`
}

type ResendRegistrationRequest struct {
	VerificationID string `json:"verification_id" binding:"required,uuid"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required,max=72"`
}

func (r LoginRequest) Identifier() string {
	if strings.TrimSpace(r.Login) != "" {
		return r.Login
	}
	return r.Email
}

type AuthResponse struct {
	Token string             `json:"token"`
	User  models.PrivateUser `json:"user"`
}

type PasswordChangeRequest struct {
	Password string `json:"password" binding:"required,min=6,max=72"`
}
