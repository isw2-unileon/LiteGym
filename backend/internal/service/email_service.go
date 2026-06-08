package service

import (
	"fmt"
	"log/slog"
	"net/smtp"
)

// EmailService defines the methods for sending emails in the application.
type EmailService interface {
	SendVerificationEmail(toEmail, username, verificationLink string) error
	SendPasswordResetEmail(toEmail, username, resetLink string) error
}

// SMTPEmailService implements EmailService using standard SMTP (e.g. Gmail).
type SMTPEmailService struct {
	host     string
	port     string
	user     string
	password string
}

// NewSMTPEmailService creates a new SMTPEmailService with the provided credentials.
func NewSMTPEmailService(host, port, user, password string) EmailService {
	return &SMTPEmailService{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

func (s *SMTPEmailService) sendMail(toEmail, subject, bodyHTML string) error {
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	from := fmt.Sprintf("From: LiteGym <%s>\r\n", s.user)
	to := fmt.Sprintf("To: %s\r\n", toEmail)
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	msg := []byte(from + to + subject + mime + bodyHTML)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, auth, s.user, []string{toEmail}, msg)
	if err != nil {
		slog.Error("failed to send email via SMTP", "error", err, "email", toEmail)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendVerificationEmail sends an email to the user with a verification link.
func (s *SMTPEmailService) SendVerificationEmail(toEmail, username, verificationLink string) error {
	subject := "Subject: Verifica tu cuenta de LiteGym\r\n"
	bodyHTML := fmt.Sprintf(`
		<h1>¡Hola %s!</h1>
		<p>Gracias por registrarte en LiteGym. Por favor, verifica tu correo electrónico haciendo clic en el siguiente enlace:</p>
		<a href="%s">Verificar cuenta</a>
		<p>Este enlace expirará en 24 horas.</p>
		<p>Si no te registraste, puedes ignorar este correo.</p>
	`, username, verificationLink)

	return s.sendMail(toEmail, subject, bodyHTML)
}

// SendPasswordResetEmail sends an email to the user with a password reset link.
func (s *SMTPEmailService) SendPasswordResetEmail(toEmail, username, resetLink string) error {
	subject := "Subject: Recupera tu contraseña de LiteGym\r\n"
	bodyHTML := fmt.Sprintf(`
		<h1>¡Hola %s!</h1>
		<p>Hemos recibido una solicitud para restablecer tu contraseña en LiteGym.</p>
		<p>Haz clic en el siguiente enlace para crear una nueva contraseña:</p>
		<a href="%s">Restablecer contraseña</a>
		<p>Este enlace expirará en 1 hora.</p>
		<p>Si no fuiste tú quien solicitó el cambio, puedes ignorar este correo de forma segura.</p>
	`, username, resetLink)

	return s.sendMail(toEmail, subject, bodyHTML)
}

// MockEmailService prints emails to the log for development.
type MockEmailService struct{}

// NewMockEmailService creates a new mock email service that prints to log.
func NewMockEmailService() EmailService {
	return &MockEmailService{}
}

// SendVerificationEmail logs the verification email instead of sending it.
func (s *MockEmailService) SendVerificationEmail(toEmail, username, verificationLink string) error {
	slog.Info("MOCK EMAIL SENT",
		"to", toEmail,
		"subject", "Verifica tu cuenta de LiteGym",
		"link", verificationLink,
	)
	return nil
}

// SendPasswordResetEmail logs the password reset email instead of sending it.
func (s *MockEmailService) SendPasswordResetEmail(toEmail, username, resetLink string) error {
	slog.Info("MOCK EMAIL SENT",
		"to", toEmail,
		"subject", "Recupera tu contraseña de LiteGym",
		"link", resetLink,
	)
	return nil
}
