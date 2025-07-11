package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/models"
	"AchinthaPallegedara/real-task/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOAuthConfig *oauth2.Config
)

// GoogleUserInfo represents the user info returned by Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// OAuthUser represents a user from OAuth providers
type OAuthUser struct {
	ID            string
	Email         string
	Name          string
	Picture       string
	EmailVerified bool
	Provider      string
}

// InitOAuthConfigs initializes OAuth configurations
func InitOAuthConfigs() {
	// Initialize Google OAuth
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GetGoogleAuthURL returns the Google OAuth authorization URL
func GetGoogleAuthURL(state string) string {
	if googleOAuthConfig == nil {
		return ""
	}
	return googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetGoogleUserInfo exchanges the authorization code for user info
func GetGoogleUserInfo(code string) (*OAuthUser, error) {
	if googleOAuthConfig == nil {
		return nil, errors.New("Google OAuth not configured")
	}

	// Exchange authorization code for token
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}

	// Get user info from Google API
	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %v", err)
	}

	return &OAuthUser{
		ID:            googleUser.ID,
		Email:         googleUser.Email,
		Name:          googleUser.Name,
		Picture:       googleUser.Picture,
		EmailVerified: googleUser.VerifiedEmail,
		Provider:      "google",
	}, nil
}

// AuthenticateOrCreateOAuthUser finds an existing user or creates a new one from OAuth
func AuthenticateOrCreateOAuthUser(oauthUser *OAuthUser) (*models.User, error) {
	var user models.User

	// Try to find existing user by email
	if err := database.DB.Where("email = ?", oauthUser.Email).First(&user).Error; err != nil {
		// User doesn't exist, create new one
		hashedPassword, err := utils.HashPassword(generateRandomPassword())
		if err != nil {
			return nil, fmt.Errorf("failed to generate password: %v", err)
		}

		user = models.User{
			Email:         oauthUser.Email,
			Password:      hashedPassword,
			Name:          oauthUser.Name,
			EmailVerified: oauthUser.EmailVerified,
			IsActive:      true,
		}

		if result := database.DB.Create(&user); result.Error != nil {
			return nil, fmt.Errorf("failed to create user: %v", result.Error)
		}

		// Send welcome email for new OAuth users
		if emailService != nil {
			// Note: Add SendWelcomeEmail method to EmailService if needed
			fmt.Printf("Welcome email would be sent to OAuth user: %s via %s\n", user.Email, oauthUser.Provider)
		}
	} else {
		// User exists, update last login
		now := time.Now()
		user.LastLoginAt = &now
		database.DB.Save(&user)
	}

	return &user, nil
}

func generateRandomPassword() string {
	// Generate a random password for OAuth users
	// They won't use this password as they'll login via OAuth
	token, _ := generateOAuthSecureToken()
	return fmt.Sprintf("oauth_%d_%s", time.Now().Unix(), token[:16])
}

// generateOAuthSecureToken generates a secure random token for OAuth service
func generateOAuthSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
