package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/config"
	"veloham/backend/internal/database"
	"veloham/backend/internal/models"
	"veloham/backend/internal/routes"
	"veloham/backend/migrations"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	db := database.Connect(cfg.DatabaseURL)
	if err := migrations.Up(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	db.Model(&models.User{}).Where("role = '' OR role IS NULL").Update("role", "user")
	if err := revokeLegacyAdmin(db); err != nil {
		log.Fatalf("revoke legacy admin: %v", err)
	}
	if err := bootstrapAdmin(db, cfg); err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}
	cleanupExpiredRecords(db)
	cleanupCtx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	go runCleanup(cleanupCtx, db)
	db.Model(&models.Listing{}).Where("deal_type = '' OR deal_type IS NULL").Update("deal_type", "продажа")
	db.Exec(`
		UPDATE listings
		SET bike_type = CASE
			WHEN category = 'Fixed Gear' THEN 'fixed'
			WHEN category = 'Road Bike' THEN 'road'
			WHEN category = 'MTB' THEN 'mtb'
			WHEN category = 'BMX' THEN 'bmx'
			ELSE bike_type
		END
		WHERE bike_type IS NULL OR bike_type = ''
	`)
	db.Exec(`UPDATE listings SET labels = '[]'::jsonb WHERE labels IS NULL`)
	db.Exec(`
		UPDATE listings
		SET
			labels = (
				SELECT COALESCE(jsonb_agg(DISTINCT label), '[]'::jsonb)
				FROM (
					SELECT jsonb_array_elements_text(labels) AS label
					UNION ALL
					SELECT 'с моей доплатой' WHERE deal_type = 'обмен с моей доплатой'
					UNION ALL
					SELECT 'с вашей доплатой' WHERE deal_type = 'обмен с доплатой покупателя'
				) AS source
				WHERE label IN ('срочно', 'торг', 'с моей доплатой', 'с вашей доплатой')
			),
			deal_type = CASE
				WHEN deal_type IN ('обмен с моей доплатой', 'обмен с доплатой покупателя') THEN 'обмен'
				WHEN deal_type IN ('продажа', 'обмен', 'продажа или обмен') THEN deal_type
				ELSE 'продажа'
			END
		WHERE deal_type NOT IN ('продажа', 'обмен', 'продажа или обмен')
	`)
	r := routes.Setup(db, cfg)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("run server: %v", err)
		}
	case <-shutdownSignal.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}
}

func cleanupExpiredRecords(db *gorm.DB) {
	now := time.Now()
	if result := db.Where("expires_at < ?", now).Delete(&models.PendingRegistration{}); result.Error != nil {
		log.Printf("cleanup expired registrations: %v", result.Error)
	}
	if result := db.Where("expires_at < ?", now).Delete(&models.PasswordReset{}); result.Error != nil {
		log.Printf("cleanup expired password resets: %v", result.Error)
	}
}

func runCleanup(ctx context.Context, db *gorm.DB) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanupExpiredRecords(db)
		}
	}
}

func revokeLegacyAdmin(db *gorm.DB) error {
	return db.Model(&models.User{}).
		Where("email = ?", "admin@veloham.kg").
		Updates(map[string]any{"role": "user", "is_blocked": true, "password_hash": "disabled-known-credential"}).Error
}

func bootstrapAdmin(db *gorm.DB, cfg config.Config) error {
	if cfg.AdminEmail == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin := models.User{
		Username:     "VELOHAM Admin",
		Email:        cfg.AdminEmail,
		PasswordHash: string(hash),
		City:         "Бишкек",
		Contact:      "@veloham_admin",
		Role:         "admin",
		IsBlocked:    false,
	}

	var existing models.User
	err = db.Where("email = ?", admin.Email).First(&existing).Error
	if err == nil {
		return db.Model(&existing).Updates(map[string]any{
			"username":      admin.Username,
			"password_hash": admin.PasswordHash,
			"city":          admin.City,
			"contact":       admin.Contact,
			"role":          admin.Role,
			"is_blocked":    admin.IsBlocked,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	return db.Create(&admin).Error
}
