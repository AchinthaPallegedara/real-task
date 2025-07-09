package services

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587 // Default SMTP port
	}

	return &EmailService{
		host:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		port:     port,
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     getEnvOrDefault("FROM_EMAIL", os.Getenv("SMTP_USERNAME")),
	}
}

// SendEmailVerification sends an email verification message
func (e *EmailService) SendEmailVerification(to, token string) error {
	subject := "Verify Your Email Address"
	verifyURL := fmt.Sprintf("%s/api/auth/verify-email?token=%s", 
		getEnvOrDefault("BASE_URL", "http://localhost:8080"), token)
	
	body := fmt.Sprintf(`
		<h2>Welcome to Task Manager!</h2>
		<p>Please click the link below to verify your email address:</p>
		<a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Verify Email</a>
		<p>If the button doesn't work, copy and paste this URL into your browser:</p>
		<p>%s</p>
		<p>This link will expire in 24 hours.</p>
		<p>If you didn't create an account, please ignore this email.</p>
	`, verifyURL, verifyURL)

	return e.sendEmail(to, subject, body)
}

// SendPasswordReset sends a password reset email
func (e *EmailService) SendPasswordReset(to, token string) error {
	subject := "Reset Your Password"
	resetURL := fmt.Sprintf("%s/api/auth/reset-password?token=%s", 
		getEnvOrDefault("BASE_URL", "http://localhost:8080"), token)
	
	body := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>You requested to reset your password. Click the link below to set a new password:</p>
		<a href="%s" style="background-color: #ff6b6b; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a>
		<p>If the button doesn't work, copy and paste this URL into your browser:</p>
		<p>%s</p>
		<p>This link will expire in 1 hour.</p>
		<p>If you didn't request this reset, please ignore this email and your password will remain unchanged.</p>
	`, resetURL, resetURL)

	return e.sendEmail(to, subject, body)
}

// SendSecurityAlert sends a security alert email
func (e *EmailService) SendSecurityAlert(to, alertType, details string) error {
	subject := fmt.Sprintf("Security Alert: %s", alertType)
	
	body := fmt.Sprintf(`
		<h2>Security Alert</h2>
		<p><strong>Alert Type:</strong> %s</p>
		<p><strong>Details:</strong> %s</p>
		<p>If this action was not performed by you, please contact support immediately.</p>
		<p>For your security, we recommend:</p>
		<ul>
			<li>Change your password immediately</li>
			<li>Enable two-factor authentication if not already enabled</li>
			<li>Review your account activity</li>
		</ul>
	`, alertType, details)

	return e.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (e *EmailService) sendEmail(to, subject, body string) error {
	// Skip email sending in development if SMTP not configured
	if e.username == "" || e.password == "" {
		fmt.Printf("Email would be sent to %s with subject: %s\n", to, subject)
		fmt.Printf("Body: %s\n", body)
		return nil // Simulate successful sending
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.host, e.port, e.username, e.password)

	return d.DialAndSend(m)
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
