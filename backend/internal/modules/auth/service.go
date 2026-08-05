package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user blocked")
	ErrEmailExists        = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	db  *gorm.DB
	jwt services.JWTService
}

func NewService(db *gorm.DB, jwt services.JWTService) *Service {
	return &Service{db: db, jwt: jwt}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	username := strings.TrimSpace(req.Username)
	if username == "" {
		username = email
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		City:         strings.TrimSpace(req.City),
		Contact:      strings.TrimSpace(req.Contact),
		Role:         "user",
	}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return AuthResponse{}, ErrEmailExists
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	var user models.User
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if user.IsBlocked {
		return AuthResponse{}, ErrUserBlocked
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user}, nil
}

func (s *Service) CurrentUser(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return models.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) AcceptPasswordRecovery(_ context.Context, _ PasswordForgotRequest) error {
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID string, req PasswordChangeRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).
		Update("password_hash", string(hash)).Error
}
