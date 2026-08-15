package config

import (
	"strings"
	"testing"
)

func TestValidateProductionSecurity(t *testing.T) {
	base := Config{
		Environment:          "production",
		JWTSecret:            "a-production-secret-with-more-than-32-characters",
		CORSOrigin:           "https://veloham.example",
		APIBaseURL:           "https://api.veloham.example",
		PublicBaseURL:        "https://veloham.example",
		FreedomPayMerchantID: "merchant-1",
		FreedomPaySecret:     "payment-secret",
		SMTPHost:             "smtp.example.com",
		SMTPUsername:         "mailer",
		SMTPPassword:         "mail-secret",
		SMTPFrom:             "VELOHAM <no-reply@example.com>",
	}
	tests := []struct {
		name   string
		mutate func(*Config)
		valid  bool
	}{
		{name: "secure production config", valid: true},
		{name: "short JWT secret", mutate: func(c *Config) { c.JWTSecret = "short" }},
		{name: "wildcard CORS", mutate: func(c *Config) { c.CORSOrigin = "*" }},
		{name: "insecure production origin", mutate: func(c *Config) { c.CORSOrigin = "http://veloham.example" }},
		{name: "insecure API URL", mutate: func(c *Config) { c.APIBaseURL = "http://api.veloham.example" }},
		{name: "admin email without password", mutate: func(c *Config) { c.AdminEmail = "admin@example.com" }},
		{name: "short admin password", mutate: func(c *Config) { c.AdminEmail = "admin@example.com"; c.AdminPassword = "short" }},
		{name: "valid admin credentials", mutate: func(c *Config) { c.AdminEmail = "admin@example.com"; c.AdminPassword = "abc123" }, valid: true},
		{name: "invalid trusted proxy", mutate: func(c *Config) { c.TrustedProxies = []string{"not-a-network"} }},
		{name: "valid trusted proxy", mutate: func(c *Config) { c.TrustedProxies = []string{"172.16.0.0/12"} }, valid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}
			err := cfg.Validate()
			if tt.valid && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !tt.valid && err == nil {
				t.Fatal("Validate() accepted insecure configuration")
			}
		})
	}
}

func TestDatabaseURLSafelyEscapesCredentials(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "veloham")
	t.Setenv("DB_PASSWORD", "p@ss:/word")
	t.Setenv("DB_NAME", "veloham")
	t.Setenv("DB_SSLMODE", "require")

	got := databaseURL()
	if !strings.Contains(got, "p%40ss%3A%2Fword") || !strings.Contains(got, "sslmode=require") {
		t.Fatalf("databaseURL() did not safely encode the DSN: %s", got)
	}
}
