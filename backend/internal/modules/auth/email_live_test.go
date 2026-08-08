package auth

import (
	"context"
	"os"
	"testing"
)

func TestLiveSMTP(t *testing.T) {
	if os.Getenv("RUN_LIVE_SMTP") != "1" {
		t.Skip("RUN_LIVE_SMTP is not enabled")
	}
	mailer := SMTPMailer{
		Host: os.Getenv("SMTP_HOST"), Port: os.Getenv("SMTP_PORT"), Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"), From: os.Getenv("SMTP_FROM"),
	}
	if err := mailer.SendVerificationCode(context.Background(), os.Getenv("SMTP_USERNAME"), "123456"); err != nil {
		t.Fatalf("send verification email: %v", err)
	}
}
