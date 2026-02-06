package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/sirupsen/logrus"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type NoopEmailSender struct{}

func (n NoopEmailSender) Send(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

type SMTPEmailSender struct {
	cfg    config.EmailConfig
	logger *logrus.Logger
}

func NewEmailSender(cfg config.EmailConfig, logger *logrus.Logger) EmailSender {
	if cfg.Host == "" || cfg.FromEmail == "" {
		if logger != nil {
			logger.Debug("SMTP not configured; using no-op email sender")
		}
		return NoopEmailSender{}
	}
	return &SMTPEmailSender{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *SMTPEmailSender) Send(_ context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	from := formatEmailAddress(s.cfg.FromName, s.cfg.FromEmail)
	msg := buildEmailMessage(from, to, subject, body)

	tlsConfig := &tls.Config{
		ServerName:         s.cfg.Host,
		InsecureSkipVerify: s.cfg.SkipTLSVerify,
	}

	var conn net.Conn
	var err error
	if s.cfg.UseTLS {
		conn, err = tls.Dial("tcp", addr, tlsConfig)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp connect failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Quit()

	if s.cfg.UseStartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("smtp starttls failed: %w", err)
			}
		} else if s.cfg.RequireStartTLS {
			return fmt.Errorf("smtp server does not support STARTTLS")
		}
	}

	if s.cfg.Username != "" || s.cfg.Password != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth failed: %w", err)
			}
		}
	}

	if err := client.Mail(s.cfg.FromEmail); err != nil {
		return fmt.Errorf("smtp mail from failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data failed: %w", err)
	}
	if _, err := writer.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}

	return nil
}

func formatEmailAddress(name, email string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return email
	}
	escaped := strings.ReplaceAll(name, "\"", "'")
	return fmt.Sprintf("\"%s\" <%s>", escaped, email)
}

func buildEmailMessage(from, to, subject, body string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}
