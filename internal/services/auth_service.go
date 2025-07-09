package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/models"
	"AchinthaPallegedara/real-task/internal/utils"
)

var (
	emailService     *EmailService
	twoFactorService *TwoFactorService
)

func init() {
	emailService = NewEmailService()
	twoFactorService = NewTwoFactorService()
}

// Constants for security settings
const (
	MaxLoginAttempts     = 5
	LockoutDuration      = 30 * time.Minute
	VerificationExpiry   = 24 * time.Hour
	PasswordResetExpiry  = 1 * time.Hour
	RefreshTokenExpiry   = 7 * 24 * time.Hour // Refresh token expires in 7 days
)

// RegisterUser creates a new user and sends verification email
func RegisterUser(email, password string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists with this email")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	verificationExpiry := time.Now().Add(VerificationExpiry)

	user := models.User{
		Email:                   email,
		Password:                hashedPassword,
		EmailVerified:          false,
		EmailVerificationToken: verificationToken,
		EmailVerificationExpiry: &verificationExpiry,
		IsActive:               true,
	}

	if result := database.DB.Create(&user); result.Error != nil {
		return nil, result.Error
	}

	// Send verification email
	if err := emailService.SendEmailVerification(email, verificationToken); err != nil {
		fmt.Printf("Failed to send verification email: %v\n", err)
		// Don't fail registration if email fails
	}

	return &user, nil
}

// AuthenticateUser verifies user credentials with security features
func AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	if result := database.DB.Where("email = ?", email).First(&user); result.Error != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, errors.New("account is temporarily locked due to too many failed attempts")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := utils.CheckPasswordHash(password, user.Password); err != nil {
		// Increment failed login attempts
		user.LoginAttempts++
		
		// Lock account if max attempts reached
		if user.LoginAttempts >= MaxLoginAttempts {
			lockUntil := time.Now().Add(LockoutDuration)
			user.LockedUntil = &lockUntil
			
			// Send security alert
			emailService.SendSecurityAlert(user.Email, "Account Locked", 
				fmt.Sprintf("Your account has been locked due to %d failed login attempts", MaxLoginAttempts))
		}
		
		database.DB.Save(&user)
		return nil, errors.New("invalid credentials")
	}

	// Reset login attempts on successful login
	user.LoginAttempts = 0
	user.LockedUntil = nil
	now := time.Now()
	user.LastLoginAt = &now
	database.DB.Save(&user)

	return &user, nil
}

// VerifyEmail verifies user's email address
func VerifyEmail(token string) error {
	var user models.User
	if err := database.DB.Where("email_verification_token = ?", token).First(&user).Error; err != nil {
		return errors.New("invalid verification token")
	}

	// Check if token is expired
	if user.EmailVerificationExpiry != nil && time.Now().After(*user.EmailVerificationExpiry) {
		return errors.New("verification token has expired")
	}

	// Mark email as verified
	user.EmailVerified = true
	user.EmailVerificationToken = ""
	user.EmailVerificationExpiry = nil

	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// ResendVerificationEmail resends verification email
func ResendVerificationEmail(email string) error {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if user.EmailVerified {
		return errors.New("email is already verified")
	}

	// Generate new verification token
	verificationToken, err := generateSecureToken()
	if err != nil {
		return err
	}

	verificationExpiry := time.Now().Add(VerificationExpiry)
	user.EmailVerificationToken = verificationToken
	user.EmailVerificationExpiry = &verificationExpiry

	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	return emailService.SendEmailVerification(email, verificationToken)
}

// RequestPasswordReset initiates password reset process
func RequestPasswordReset(email string) error {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate reset token
	resetToken, err := generateSecureToken()
	if err != nil {
		return err
	}

	resetExpiry := time.Now().Add(PasswordResetExpiry)
	user.PasswordResetToken = resetToken
	user.PasswordResetExpiry = &resetExpiry

	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	return emailService.SendPasswordReset(email, resetToken)
}

// ResetPassword resets user password using token
func ResetPassword(token, newPassword string) error {
	var user models.User
	if err := database.DB.Where("password_reset_token = ?", token).First(&user).Error; err != nil {
		return errors.New("invalid reset token")
	}

	// Check if token is expired
	if user.PasswordResetExpiry != nil && time.Now().After(*user.PasswordResetExpiry) {
		return errors.New("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	user.Password = hashedPassword
	user.PasswordResetToken = ""
	user.PasswordResetExpiry = nil
	user.LoginAttempts = 0 // Reset failed attempts
	user.LockedUntil = nil // Unlock account

	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	// Send security alert
	emailService.SendSecurityAlert(user.Email, "Password Changed", 
		"Your password has been successfully changed")

	return nil
}

// SetupTwoFactor generates TOTP secret for user
func SetupTwoFactor(userID uint) (string, string, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return "", "", errors.New("user not found")
	}

	if user.TwoFactorEnabled {
		return "", "", errors.New("two-factor authentication is already enabled")
	}

	// Generate TOTP secret
	secret, err := twoFactorService.GenerateSecret(user.Email)
	if err != nil {
		return "", "", err
	}

	// Generate QR code URL
	qrCode, err := twoFactorService.GenerateQRCode(user.Email, secret)
	if err != nil {
		return "", "", err
	}

	// Store secret temporarily (will be confirmed when user verifies)
	user.TwoFactorSecret = secret
	if err := database.DB.Save(&user).Error; err != nil {
		return "", "", err
	}

	return secret, qrCode, nil
}

// ConfirmTwoFactor enables 2FA after token verification
func ConfirmTwoFactor(userID uint, token string) ([]string, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if user.TwoFactorSecret == "" {
		return nil, errors.New("two-factor setup not initiated")
	}

	// Validate the token
	if !twoFactorService.ValidateToken(user.TwoFactorSecret, token) {
		return nil, errors.New("invalid verification token")
	}

	// Generate backup codes
	backupCodes, err := twoFactorService.GenerateBackupCodes()
	if err != nil {
		return nil, err
	}

	// Enable 2FA
	user.TwoFactorEnabled = true
	user.TwoFactorBackupCodes = models.StringSlice(backupCodes)

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	// Send security alert
	emailService.SendSecurityAlert(user.Email, "Two-Factor Authentication Enabled", 
		"Two-factor authentication has been successfully enabled on your account")

	return backupCodes, nil
}

// DisableTwoFactor disables 2FA for user
func DisableTwoFactor(userID uint, password string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify password before disabling 2FA
	if err := utils.CheckPasswordHash(password, user.Password); err != nil {
		return errors.New("invalid password")
	}

	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""
	user.TwoFactorBackupCodes = models.StringSlice(nil)

	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	// Send security alert
	emailService.SendSecurityAlert(user.Email, "Two-Factor Authentication Disabled", 
		"Two-factor authentication has been disabled on your account")

	return nil
}

// VerifyTwoFactor verifies 2FA token or backup code
func VerifyTwoFactor(userID uint, token string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if !user.TwoFactorEnabled {
		return errors.New("two-factor authentication is not enabled")
	}

	// Try TOTP token first
	if twoFactorService.ValidateToken(user.TwoFactorSecret, token) {
		return nil
	}

	// Try backup codes
	valid, remainingCodes := twoFactorService.ValidateBackupCode([]string(user.TwoFactorBackupCodes), token)
	if valid {
		user.TwoFactorBackupCodes = models.StringSlice(remainingCodes)
		database.DB.Save(&user)
		return nil
	}

	return errors.New("invalid verification token")
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateTokenPair creates both access and refresh tokens for a user
func GenerateTokenPair(userID uint) (*utils.TokenPair, error) {
	tokenPair, err := utils.GenerateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	refreshExpiry := time.Now().Add(RefreshTokenExpiry)
	user.RefreshToken = tokenPair.RefreshToken
	user.RefreshTokenExpiry = &refreshExpiry

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func RefreshAccessToken(refreshToken string) (*utils.TokenPair, error) {
	var user models.User
	if err := database.DB.Where("refresh_token = ?", refreshToken).First(&user).Error; err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if refresh token is expired
	if user.RefreshTokenExpiry == nil || time.Now().After(*user.RefreshTokenExpiry) {
		return nil, errors.New("refresh token has expired")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new token pair
	return GenerateTokenPair(user.ID)
}

// RevokeRefreshToken invalidates a refresh token
func RevokeRefreshToken(userID uint) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	user.RefreshToken = ""
	user.RefreshTokenExpiry = nil

	return database.DB.Save(&user).Error
}

// RevokeAllUserTokens revokes all refresh tokens for a user (useful for logout all devices)
func RevokeAllUserTokens(userID uint) error {
	return RevokeRefreshToken(userID) // Currently we only store one refresh token per user
}