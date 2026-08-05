package auth

import "veloham/backend/internal/models"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	City     string `json:"city"`
	Contact  string `json:"contact"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type PasswordForgotRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordChangeRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}
