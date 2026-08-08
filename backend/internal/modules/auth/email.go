package auth

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

type SMTPMailer struct {
	Host, Port, Username, Password, From string
}

func (m SMTPMailer) SendVerificationCode(ctx context.Context, recipient, code string) error {
	return m.sendCode(ctx, recipient, code, "VELOHAM - код подтверждения", "Ваш код подтверждения VELOHAM: ")
}

func (m SMTPMailer) SendPasswordResetCode(ctx context.Context, recipient, code string) error {
	return m.sendCode(ctx, recipient, code, "VELOHAM - восстановление пароля", "Код для восстановления пароля VELOHAM: ")
}

func (m SMTPMailer) sendCode(ctx context.Context, recipient, code, subject, intro string) error {
	from, err := mail.ParseAddress(m.From)
	if err != nil {
		return fmt.Errorf("invalid SMTP_FROM: %w", err)
	}
	to, err := mail.ParseAddress(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	address := net.JoinHostPort(m.Host, m.Port)
	connection, err := (&net.Dialer{Timeout: 12 * time.Second}).DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(connection, m.Host)
	if err != nil {
		_ = connection.Close()
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return errors.New("SMTP server does not support STARTTLS")
	}
	if err := client.StartTLS(&tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12}); err != nil {
		return err
	}
	if err := client.Auth(smtp.PlainAuth("", m.Username, m.Password, m.Host)); err != nil {
		return err
	}
	if err := client.Mail(from.Address); err != nil {
		return err
	}
	if err := client.Rcpt(to.Address); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	message := strings.Join([]string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		intro + code,
		"Код действует 10 минут. Никому его не сообщайте.",
	}, "\r\n")
	if _, err := io.WriteString(writer, message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func randomNumericCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

var errEmailNotConfigured = errors.New("email delivery is not configured")
