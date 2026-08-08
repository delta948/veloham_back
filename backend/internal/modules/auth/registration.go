package auth

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"veloham/backend/internal/models"
)

const (
	registrationTTL = 10 * time.Minute
	resendDelay     = time.Minute
	maxCodeAttempts = 5
	maxCodeResends  = 5
)

var kyrgyzPhonePattern = regexp.MustCompile(`^\+996[0-9]{9}$`)

func normalizeKyrgyzPhone(raw string) (string, bool) {
	phone := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	if strings.HasPrefix(phone, "0") && len(phone) == 10 {
		phone = "+996" + phone[1:]
	} else if strings.HasPrefix(phone, "996") {
		phone = "+" + phone
	}
	return phone, kyrgyzPhonePattern.MatchString(phone)
}

func normalizedLoginPhone(raw string) string {
	phone, valid := normalizeKyrgyzPhone(raw)
	if !valid {
		return ""
	}
	return phone
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegistrationStartResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	username := strings.TrimSpace(req.Username)
	var existing int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&existing).Error; err != nil {
		return RegistrationStartResponse{}, err
	}
	if existing > 0 {
		return RegistrationStartResponse{}, ErrEmailExists
	}
	var pending models.PendingRegistration
	pendingErr := s.db.WithContext(ctx).Where("email = ?", email).First(&pending).Error
	if pendingErr != nil && !errors.Is(pendingErr, gorm.ErrRecordNotFound) {
		return RegistrationStartResponse{}, pendingErr
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegistrationStartResponse{}, err
	}
	now := time.Now()
	if pendingErr == nil && now.Before(pending.ExpiresAt) {
		if now.Before(pending.ResendAfter) {
			return RegistrationStartResponse{}, ErrResendCooldown
		}
		if pending.ResendCount >= maxCodeResends {
			return RegistrationStartResponse{}, ErrVerificationLimit
		}
		providerToken, codeHash, devCode, err := s.sendRegistrationCode(ctx, email)
		if err != nil {
			return RegistrationStartResponse{}, err
		}
		if err := s.db.WithContext(ctx).Model(&pending).Updates(map[string]any{
			"username": username, "password_hash": string(passwordHash), "city": strings.TrimSpace(req.City), "contact": strings.TrimSpace(req.Contact),
			"provider_token": providerToken, "code_hash": codeHash, "attempts": 0,
			"resend_count": gorm.Expr("resend_count + 1"), "expires_at": now.Add(registrationTTL), "resend_after": now.Add(resendDelay),
		}).Error; err != nil {
			return RegistrationStartResponse{}, err
		}
		return RegistrationStartResponse{VerificationID: pending.ID, Email: email, ExpiresIn: int(registrationTTL.Seconds()), DevCode: devCode}, nil
	}
	providerToken, codeHash, devCode, err := s.sendRegistrationCode(ctx, email)
	if err != nil {
		return RegistrationStartResponse{}, err
	}
	pending = models.PendingRegistration{
		Username: username, Email: email, City: strings.TrimSpace(req.City), Contact: strings.TrimSpace(req.Contact),
		PasswordHash: string(passwordHash), ProviderToken: providerToken, CodeHash: codeHash,
		ExpiresAt: now.Add(registrationTTL), ResendAfter: now.Add(resendDelay),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ?", email).Delete(&models.PendingRegistration{}).Error; err != nil {
			return err
		}
		return tx.Create(&pending).Error
	})
	if err != nil {
		return RegistrationStartResponse{}, err
	}
	return RegistrationStartResponse{VerificationID: pending.ID, Email: email, ExpiresIn: int(registrationTTL.Seconds()), DevCode: devCode}, nil
}

func (s *Service) ResendRegistration(ctx context.Context, verificationID string) (RegistrationStartResponse, error) {
	var pending models.PendingRegistration
	if err := s.db.WithContext(ctx).First(&pending, "id = ?", verificationID).Error; err != nil || time.Now().After(pending.ExpiresAt) {
		return RegistrationStartResponse{}, ErrVerification
	}
	if time.Now().Before(pending.ResendAfter) {
		return RegistrationStartResponse{}, ErrResendCooldown
	}
	if pending.ResendCount >= maxCodeResends {
		return RegistrationStartResponse{}, ErrVerificationLimit
	}
	providerToken, codeHash, devCode, err := s.sendRegistrationCode(ctx, pending.Email)
	if err != nil {
		return RegistrationStartResponse{}, err
	}
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&pending).Updates(map[string]any{
		"provider_token": providerToken, "code_hash": codeHash, "attempts": 0,
		"resend_count": gorm.Expr("resend_count + 1"), "expires_at": now.Add(registrationTTL), "resend_after": now.Add(resendDelay),
	}).Error; err != nil {
		return RegistrationStartResponse{}, err
	}
	return RegistrationStartResponse{VerificationID: pending.ID, Email: pending.Email, ExpiresIn: int(registrationTTL.Seconds()), DevCode: devCode}, nil
}

func (s *Service) VerifyRegistration(ctx context.Context, verificationID, code string) (AuthResponse, error) {
	var result AuthResponse
	var verificationErr error
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pending models.PendingRegistration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pending, "id = ?", verificationID).Error; err != nil {
			return ErrVerification
		}
		if time.Now().After(pending.ExpiresAt) {
			return ErrVerification
		}
		if pending.Attempts >= maxCodeAttempts {
			return ErrVerificationLimit
		}
		if err := tx.Model(&pending).UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error; err != nil {
			return err
		}
		valid := bcrypt.CompareHashAndPassword([]byte(pending.CodeHash), []byte(strings.TrimSpace(code))) == nil
		if !valid {
			verificationErr = ErrVerification
			if pending.Attempts+1 >= maxCodeAttempts {
				verificationErr = ErrVerificationLimit
			}
			return nil
		}
		user := models.User{Username: pending.Username, Email: pending.Email, PasswordHash: pending.PasswordHash, City: pending.City, Contact: pending.Contact, Role: "user"}
		if err := tx.Create(&user).Error; err != nil {
			return ErrEmailExists
		}
		if err := tx.Delete(&pending).Error; err != nil {
			return err
		}
		token, err := s.jwt.Generate(user.ID)
		if err != nil {
			return err
		}
		result = AuthResponse{Token: token, User: models.UserWithEmail(user)}
		return nil
	})
	if err == nil && verificationErr != nil {
		return AuthResponse{}, verificationErr
	}
	return result, err
}

func (s *Service) sendRegistrationCode(ctx context.Context, email string) (providerToken, codeHash, devCode string, err error) {
	code, err := randomNumericCode()
	if err != nil {
		return "", "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", "", "", err
	}
	if s.cfg.SMTPHost != "" && s.cfg.SMTPUsername != "" && s.cfg.SMTPPassword != "" && s.cfg.SMTPFrom != "" {
		if err := s.mailer.SendVerificationCode(ctx, email, code); err != nil {
			return "", "", "", err
		}
		return "email", string(hash), "", nil
	}
	if strings.EqualFold(s.cfg.Environment, "test") {
		log.Printf("test email verification code for %s: %s", email, code)
		return "test", string(hash), code, nil
	}
	return "", "", "", errEmailNotConfigured
}
