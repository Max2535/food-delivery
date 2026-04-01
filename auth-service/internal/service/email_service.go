package service

import (
	"fmt"
	"net/smtp"
	"os"
)

// EmailSender abstracts email delivery so the real SMTP impl can be swapped
// with a no-op or mock in tests.
type EmailSender interface {
	SendVerificationEmail(toEmail, username, token string) error
	SendPasswordResetEmail(toEmail, username, token string) error
}

// ── SMTP implementation ────────────────────────────────────────────────────

type smtpEmailSender struct {
	host     string
	port     string
	from     string
	password string
	siteURL  string
}

func NewSmtpEmailSender() EmailSender {
	return &smtpEmailSender{
		host:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		port:     getEnvOrDefault("SMTP_PORT", "587"),
		from:     getEnvOrDefault("SMTP_FROM", "suppchai1992@gmail.com"),
		password: os.Getenv("SMTP_PASSWORD"),
		siteURL:  getEnvOrDefault("SITE_URL", "http://localhost:3000"),
	}
}

func (s *smtpEmailSender) SendPasswordResetEmail(toEmail, username, token string) error {
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.siteURL, token)

	subject := "รีเซ็ตรหัสผ่านของคุณ — Food Delivery"
	body := fmt.Sprintf(`สวัสดี %s,

เราได้รับคำขอรีเซ็ตรหัสผ่านสำหรับบัญชีของคุณ กรุณาคลิกลิงก์ด้านล่างเพื่อตั้งรหัสผ่านใหม่ ลิงก์จะหมดอายุภายใน 15 นาที

%s

หากคุณไม่ได้ขอรีเซ็ตรหัสผ่าน กรุณาเพิกเฉยต่ออีเมลนี้

ขอบคุณ,
ทีม Food Delivery`, username, resetURL)

	msg := buildMimeMessage(s.from, toEmail, subject, body)
	auth := smtp.PlainAuth("", s.from, s.password, s.host)
	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{toEmail}, msg)
}

func (s *smtpEmailSender) SendVerificationEmail(toEmail, username, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.siteURL, token)

	subject := "ยืนยันอีเมลของคุณ — Food Delivery"
	body := fmt.Sprintf(`สวัสดี %s,

กรุณาคลิกลิงก์ด้านล่างเพื่อยืนยันอีเมลของคุณ ลิงก์จะหมดอายุภายใน 24 ชั่วโมง

%s

หากคุณไม่ได้สมัครสมาชิก กรุณาเพิกเฉยต่ออีเมลนี้

ขอบคุณ,
ทีม Food Delivery`, username, verifyURL)

	msg := buildMimeMessage(s.from, toEmail, subject, body)

	auth := smtp.PlainAuth("", s.from, s.password, s.host)
	addr := s.host + ":" + s.port
	return smtp.SendMail(addr, auth, s.from, []string{toEmail}, msg)
}

func buildMimeMessage(from, to, subject, body string) []byte {
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body,
	)
	return []byte(msg)
}

// ── No-op implementation (used when SMTP is not configured) ───────────────

type noopEmailSender struct{}

func NewNoopEmailSender() EmailSender { return &noopEmailSender{} }

func (n *noopEmailSender) SendVerificationEmail(_, _, _ string) error  { return nil }
func (n *noopEmailSender) SendPasswordResetEmail(_, _, _ string) error { return nil }

// ── Helper ────────────────────────────────────────────────────────────────

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
