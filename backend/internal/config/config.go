package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Environment          string
	Port                 string
	DatabaseURL          string
	RedisURL             string
	JWTSecret            string
	CORSOrigin           string
	TrustedProxies       []string
	UploadDir            string
	AdminEmail           string
	AdminPassword        string
	PublicBaseURL        string
	FreedomPayMerchantID string
	FreedomPaySecret     string
	FreedomPayAPIBase    string
	SMTPHost             string
	SMTPPort             string
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
}

func Load() Config {
	return Config{
		Environment:          value("APP_ENV", "development"),
		Port:                 value("APP_PORT", value("PORT", "8080")),
		DatabaseURL:          databaseURL(),
		RedisURL:             value("REDIS_URL", "redis://127.0.0.1:6379/0"),
		JWTSecret:            value("JWT_SECRET", "dev-secret"),
		CORSOrigin:           value("CORS_ORIGIN", "http://127.0.0.1:5173,http://localhost:5173"),
		TrustedProxies:       splitList(os.Getenv("TRUSTED_PROXIES")),
		UploadDir:            value("UPLOAD_DIR", "uploads"),
		AdminEmail:           strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL"))),
		AdminPassword:        os.Getenv("ADMIN_PASSWORD"),
		PublicBaseURL:        value("PUBLIC_BASE_URL", "http://127.0.0.1:5173"),
		FreedomPayMerchantID: strings.TrimSpace(os.Getenv("FREEDOMPAY_MERCHANT_ID")),
		FreedomPaySecret:     os.Getenv("FREEDOMPAY_SECRET"),
		FreedomPayAPIBase:    value("FREEDOMPAY_API_BASE", "https://api.freedompay.kg"),
		SMTPHost:             strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:             value("SMTP_PORT", "587"),
		SMTPUsername:         strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:         os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:             strings.TrimSpace(os.Getenv("SMTP_FROM")),
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
		for _, origin := range splitList(c.CORSOrigin) {
			parsed, err := url.Parse(origin)
			if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
				return errors.New("CORS_ORIGIN must contain valid HTTPS origins in production")
			}
		}
		publicURL, err := url.Parse(c.PublicBaseURL)
		if err != nil || publicURL.Scheme != "https" || publicURL.Host == "" {
			return errors.New("PUBLIC_BASE_URL must be a valid HTTPS URL in production")
		}
		if c.FreedomPayMerchantID == "" || c.FreedomPaySecret == "" {
			return errors.New("FREEDOMPAY_MERCHANT_ID and FREEDOMPAY_SECRET are required in production")
		}
		if c.SMTPHost == "" || c.SMTPUsername == "" || c.SMTPPassword == "" || c.SMTPFrom == "" {
			return errors.New("SMTP_HOST, SMTP_USERNAME, SMTP_PASSWORD and SMTP_FROM are required in production")
		}
	}
	if (c.AdminEmail == "") != (c.AdminPassword == "") {
		return errors.New("ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}
	if c.AdminPassword != "" && len(c.AdminPassword) < 6 {
		return errors.New("ADMIN_PASSWORD must contain at least 6 characters")
	}
	for _, proxy := range c.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return errors.New("TRUSTED_PROXIES must contain only IP addresses or CIDR ranges")
			}
		}
	}
	return nil
}

func splitList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
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
	dsn := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + name,
		RawQuery: url.Values{"sslmode": []string{sslMode}}.Encode(),
	}
	return dsn.String()
}
