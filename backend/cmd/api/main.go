package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"veloham/backend/internal/config"
	"veloham/backend/internal/database"
	"veloham/backend/internal/models"
	"veloham/backend/internal/routes"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	db := database.Connect(cfg.DatabaseURL)
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Fatalf("enable uuid-ossp extension: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Listing{}, &models.ListingImage{}, &models.BuildCard{}, &models.Favorite{}, &models.MatchPreference{}, &models.Chat{}, &models.Message{}, &models.Review{}, &models.Report{}, &models.WantedRequest{}, &models.WantedOffer{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	db.Model(&models.User{}).Where("role = '' OR role IS NULL").Update("role", "user")
	if err := revokeLegacyAdmin(db); err != nil {
		log.Fatalf("revoke legacy admin: %v", err)
	}
	if err := bootstrapAdmin(db, cfg); err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}
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
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
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
