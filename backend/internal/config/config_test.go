package config

import "testing"

func TestValidateProductionSecurity(t *testing.T) {
	base := Config{
		Environment: "production",
		JWTSecret:   "a-production-secret-with-more-than-32-characters",
		CORSOrigin:  "https://veloham.example",
	}
	tests := []struct {
		name   string
		mutate func(*Config)
		valid  bool
	}{
		{name: "secure production config", valid: true},
		{name: "short JWT secret", mutate: func(c *Config) { c.JWTSecret = "short" }},
		{name: "wildcard CORS", mutate: func(c *Config) { c.CORSOrigin = "*" }},
		{name: "admin email without password", mutate: func(c *Config) { c.AdminEmail = "admin@example.com" }},
		{name: "short admin password", mutate: func(c *Config) { c.AdminEmail = "admin@example.com"; c.AdminPassword = "short" }},
		{name: "valid admin credentials", mutate: func(c *Config) { c.AdminEmail = "admin@example.com"; c.AdminPassword = "long-admin-password" }, valid: true},
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
