package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTGenerateAndParse(t *testing.T) {
	service := NewJWTService("a-test-secret-that-is-long-enough-for-tests")
	token, err := service.Generate("user-123")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	userID, err := service.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("Parse() userID = %q, want user-123", userID)
	}
}

func TestJWTParseRejectsInvalidClaimsAndAlgorithms(t *testing.T) {
	secret := "a-test-secret-that-is-long-enough-for-tests"
	service := NewJWTService(secret)
	tests := []struct {
		name   string
		method jwt.SigningMethod
		claims jwt.MapClaims
	}{
		{
			name:   "expired",
			method: jwt.SigningMethodHS256,
			claims: jwt.MapClaims{"sub": "user-123", "iat": time.Now().Add(-2 * time.Hour).Unix(), "exp": time.Now().Add(-time.Hour).Unix()},
		},
		{
			name:   "missing expiration",
			method: jwt.SigningMethodHS256,
			claims: jwt.MapClaims{"sub": "user-123", "iat": time.Now().Unix()},
		},
		{
			name:   "wrong algorithm",
			method: jwt.SigningMethodHS512,
			claims: jwt.MapClaims{"sub": "user-123", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(tt.method, tt.claims)
			signed, err := token.SignedString([]byte(secret))
			if err != nil {
				t.Fatalf("SignedString() error = %v", err)
			}
			if _, err := service.Parse(signed); err == nil {
				t.Fatal("Parse() accepted an invalid token")
			}
		})
	}
}
