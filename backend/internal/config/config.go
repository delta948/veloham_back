package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	CORSOrigin  string
	UploadDir   string
}

func Load() Config {
	return Config{
		Port:        value("APP_PORT", value("PORT", "8080")),
		DatabaseURL: databaseURL(),
		RedisURL:    value("REDIS_URL", "redis://127.0.0.1:6379/0"),
		JWTSecret:   value("JWT_SECRET", "dev-secret"),
		CORSOrigin:  value("CORS_ORIGIN", "http://127.0.0.1:5173,http://localhost:5173"),
		UploadDir:   value("UPLOAD_DIR", "uploads"),
	}
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
