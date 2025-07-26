package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"AchinthaPallegedara/real-task/internal/services"

	"github.com/gin-gonic/gin"
)

// OAuth request/response structures
type OAuthLoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// GenerateOAuthState generates a secure state parameter for OAuth
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GoogleAuthURL returns the Google OAuth authorization URL
// @Summary Get Google OAuth URL
// @Description Get Google OAuth authorization URL for login
// @Tags OAuth
// @Accept json
// @Produce json
// @Success 200 {object} OAuthURLResponse "OAuth URL generated successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/auth/google/url [get]
func GoogleAuthURL(c *gin.Context) {
	state, err := generateOAuthState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state parameter"})
		return
	}

	// Store state in session/cache for verification (simplified here)
	// In production, you should store this securely
	authURL := services.GetGoogleAuthURL(state)
	if authURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth not configured"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GoogleAuthCallback handles the callback from Google OAuth
func GoogleAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	// In production, verify the state parameter against stored value
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter is required"})
		return
	}

	// Exchange code for user info
	oauthUser, err := services.GetGoogleUserInfo(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// Authenticate or create user
	user, err := services.AuthenticateOrCreateOAuthUser(oauthUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate user: " + err.Error()})
		return
	}

	// Generate JWT token
	tokenPair, err := services.GenerateTokenPair(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
		"token_type":    "Bearer",
		"user": gin.H{
			"id":                 user.ID,
			"email":             user.Email,
			"name":              user.Name,
			"email_verified":    user.EmailVerified,
			"two_factor_enabled": user.TwoFactorEnabled,
			"provider":          "google",
		},
	})
}

// GoogleAuthLogin handles POST request for Google OAuth (when using code from frontend)
func GoogleAuthLogin(c *gin.Context) {
	var req OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In production, verify the state parameter
	
	// Exchange code for user info
	oauthUser, err := services.GetGoogleUserInfo(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// Authenticate or create user
	user, err := services.AuthenticateOrCreateOAuthUser(oauthUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate user: " + err.Error()})
		return
	}

	// Generate JWT token
	tokenPair, err := services.GenerateTokenPair(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
		"token_type":    "Bearer",
		"user": gin.H{
			"id":                 user.ID,
			"email":             user.Email,
			"name":              user.Name,
			"email_verified":    user.EmailVerified,
			"two_factor_enabled": user.TwoFactorEnabled,
			"provider":          "google",
		},
	})
}

// OAuthProviders returns available OAuth providers
// @Summary Get OAuth providers
// @Description Get list of available OAuth providers
// @Tags OAuth
// @Accept json
// @Produce json
// @Success 200 {object} OAuthProvidersResponse "OAuth providers retrieved successfully"
// @Router /api/auth/providers [get]
func OAuthProviders(c *gin.Context) {
	providers := []gin.H{
		{
			"name":     "Google",
			"id":       "google",
			"endpoint": "/api/auth/google",
			"enabled":  services.GetGoogleAuthURL("test") != "",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}
