package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/config"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user blocked")
	ErrEmailExists        = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrVerification       = errors.New("invalid or expired verification code")
	ErrVerificationLimit  = errors.New("too many verification attempts")
	ErrResendCooldown     = errors.New("wait before requesting another code")
)

type UserBlockedError struct {
	Reason string
}

func (e UserBlockedError) Error() string        { return ErrUserBlocked.Error() }
func (e UserBlockedError) Is(target error) bool { return target == ErrUserBlocked }

type Service struct {
	db     *gorm.DB
	jwt    services.JWTService
	cfg    config.Config
	mailer SMTPMailer
}

func NewService(db *gorm.DB, jwt services.JWTService, cfg config.Config) *Service {
	return &Service{db: db, jwt: jwt, cfg: cfg, mailer: SMTPMailer{Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUsername, Password: cfg.SMTPPassword, From: cfg.SMTPFrom}}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	var user models.User
	identifier := strings.TrimSpace(strings.ToLower(req.Identifier()))
	query := s.db.WithContext(ctx).Where("LOWER(email) = ? OR LOWER(username) = ?", identifier, identifier)
	if phone := normalizedLoginPhone(identifier); phone != "" {
		query = query.Or("phone = ?", phone)
	}
	if err := query.First(&user).Error; err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if user.IsBlocked {
		return AuthResponse{}, UserBlockedError{Reason: user.BlockedReason}
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: models.UserWithEmail(user)}, nil
}

func (s *Service) CurrentUser(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return models.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID string, req PasswordChangeRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).
		Update("password_hash", string(hash)).Error
}
