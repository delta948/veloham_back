package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Environment   string
	Port          string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	CORSOrigin    string
	UploadDir     string
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	return Config{
		Environment:   value("APP_ENV", "development"),
		Port:          value("APP_PORT", value("PORT", "8080")),
		DatabaseURL:   databaseURL(),
		RedisURL:      value("REDIS_URL", "redis://127.0.0.1:6379/0"),
		JWTSecret:     value("JWT_SECRET", "dev-secret"),
		CORSOrigin:    value("CORS_ORIGIN", "http://127.0.0.1:5173,http://localhost:5173"),
		UploadDir:     value("UPLOAD_DIR", "uploads"),
		AdminEmail:    strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL"))),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}
}

func (c Config) Validate() error {
	if strings.EqualFold(c.Environment, "production") {
		if len(c.JWTSecret) < 32 || c.JWTSecret == "dev-secret" || c.JWTSecret == "change-me-in-production" {
			return errors.New("JWT_SECRET must contain at least 32 characters in production")
		}
		if strings.Contains(c.CORSOrigin, "*") {
			return errors.New("CORS_ORIGIN must list explicit origins in production")
		}
	}
	if (c.AdminEmail == "") != (c.AdminPassword == "") {
		return errors.New("ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}
	if c.AdminPassword != "" && len(c.AdminPassword) < 12 {
		return errors.New("ADMIN_PASSWORD must contain at least 12 characters")
	}
	return nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func databaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	host := value("DB_HOST", "127.0.0.1")
	port := value("DB_PORT", "5432")
	user := value("DB_USER", "veloham")
	password := value("DB_PASSWORD", "veloham")
	name := value("DB_NAME", "veloham")
	sslMode := value("DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslMode)
}
