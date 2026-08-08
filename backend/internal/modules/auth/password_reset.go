package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"veloham/backend/internal/models"
)

func (s *Service) ForgotPassword(ctx context.Context, email string) (ForgotPasswordResponse, error) {
	fakeID, err := randomUUID()
	if err != nil {
		return ForgotPasswordResponse{}, err
	}
	var user models.User
	if err := s.db.WithContext(ctx).Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ForgotPasswordResponse{ResetID: fakeID, ExpiresIn: int(registrationTTL.Seconds())}, nil
		}
		return ForgotPasswordResponse{}, err
	}
	var existing models.PasswordReset
	if err := s.db.WithContext(ctx).Where("user_id = ?", user.ID).Order("created_at desc").First(&existing).Error; err == nil && time.Now().Before(existing.ResendAfter) {
		return ForgotPasswordResponse{ResetID: existing.ID, ExpiresIn: int(time.Until(existing.ExpiresAt).Seconds())}, nil
	}
	code, err := randomNumericCode()
	if err != nil {
		return ForgotPasswordResponse{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return ForgotPasswordResponse{}, err
	}
	devCode := ""
	if s.cfg.SMTPHost != "" && s.cfg.SMTPUsername != "" && s.cfg.SMTPPassword != "" && s.cfg.SMTPFrom != "" {
		if err := s.mailer.SendPasswordResetCode(ctx, user.Email, code); err != nil {
			return ForgotPasswordResponse{}, err
		}
	} else if strings.EqualFold(s.cfg.Environment, "test") {
		log.Printf("test password reset code for %s: %s", user.Email, code)
		devCode = code
	} else {
		return ForgotPasswordResponse{}, errEmailNotConfigured
	}
	now := time.Now()
	reset := models.PasswordReset{UserID: user.ID, CodeHash: string(hash), ExpiresAt: now.Add(registrationTTL), ResendAfter: now.Add(resendDelay)}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.PasswordReset{}).Error; err != nil {
			return err
		}
		return tx.Create(&reset).Error
	}); err != nil {
		return ForgotPasswordResponse{}, err
	}
	return ForgotPasswordResponse{ResetID: reset.ID, ExpiresIn: int(registrationTTL.Seconds()), DevCode: devCode}, nil
}

func randomUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func (s *Service) ResetPassword(ctx context.Context, resetID, code, password string) error {
	var verificationErr error
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reset models.PasswordReset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&reset, "id = ?", resetID).Error; err != nil || time.Now().After(reset.ExpiresAt) {
			return ErrVerification
		}
		if reset.Attempts >= maxCodeAttempts {
			return ErrVerificationLimit
		}
		if bcrypt.CompareHashAndPassword([]byte(reset.CodeHash), []byte(strings.TrimSpace(code))) != nil {
			if err := tx.Model(&reset).UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error; err != nil {
				return err
			}
			verificationErr = ErrVerification
			if reset.Attempts+1 >= maxCodeAttempts {
				verificationErr = ErrVerificationLimit
			}
			return nil
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", reset.UserID).Update("password_hash", string(hash)).Error; err != nil {
			return err
		}
		return tx.Delete(&reset).Error
	})
	if err == nil && verificationErr != nil {
		return verificationErr
	}
	return err
}
