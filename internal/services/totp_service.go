package services

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TwoFactorService struct{}

// NewTwoFactorService creates a new 2FA service instance
func NewTwoFactorService() *TwoFactorService {
	return &TwoFactorService{}
}

// GenerateSecret generates a new TOTP secret for a user
func (t *TwoFactorService) GenerateSecret(userEmail string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Task Manager",
		AccountName: userEmail,
		SecretSize:  32,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

// GenerateQRCode generates QR code URL for TOTP setup
func (t *TwoFactorService) GenerateQRCode(userEmail, secret string) (string, error) {
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/Task%%20Manager:%s?secret=%s&issuer=Task%%20Manager", userEmail, secret))
	if err != nil {
		return "", err
	}
	return key.URL(), nil
}

// ValidateToken validates a TOTP token
func (t *TwoFactorService) ValidateToken(secret, token string) bool {
	return totp.Validate(token, secret)
}

// GenerateBackupCodes generates backup codes for 2FA
func (t *TwoFactorService) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, 10) // Generate 10 backup codes
	
	for i := 0; i < 10; i++ {
		code, err := t.generateRandomCode(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	
	return codes, nil
}

// ValidateBackupCode validates a backup code
func (t *TwoFactorService) ValidateBackupCode(codes []string, inputCode string) (bool, []string) {
	inputCode = strings.ToUpper(strings.TrimSpace(inputCode))
	
	for i, code := range codes {
		if code == inputCode {
			// Remove the used backup code
			newCodes := make([]string, len(codes)-1)
			copy(newCodes, codes[:i])
			copy(newCodes[i:], codes[i+1:])
			return true, newCodes
		}
	}
	
	return false, codes
}

// generateRandomCode generates a random alphanumeric code
func (t *TwoFactorService) generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	
	return string(bytes), nil
}

// EncodeSecret encodes a secret for storage
func (t *TwoFactorService) EncodeSecret(secret string) string {
	return base32.StdEncoding.EncodeToString([]byte(secret))
}

// DecodeSecret decodes a secret from storage
func (t *TwoFactorService) DecodeSecret(encodedSecret string) (string, error) {
	decoded, err := base32.StdEncoding.DecodeString(encodedSecret)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
