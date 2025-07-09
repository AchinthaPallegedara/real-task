package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GenerateToken generates a JWT token for a given user ID (legacy function for backward compatibility)
func GenerateToken(userID uint) (string, error) {
	return GenerateAccessToken(userID)
}

// GenerateAccessToken generates a short-lived access token
func GenerateAccessToken(userID uint) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable not set")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // Access token expires in 1 hour
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GenerateRefreshToken generates a long-lived refresh token
func GenerateRefreshToken() (string, error) {
	// Generate a secure random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uint) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 1), // Access token expiry
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func ValidateToken(tokenString string) (uint, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return 0, fmt.Errorf("JWT_SECRET environment variable not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check if it's an access token
		if tokenType, exists := claims["type"]; exists && tokenType != "access" {
			return 0, fmt.Errorf("invalid token type")
		}

		userID := uint(claims["user_id"].(float64)) // JWT numerical claims are float64
		return userID, nil
	}
	return 0, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates a refresh token from the database
func ValidateRefreshToken(refreshToken string) (uint, error) {
	// This function will be used by the auth service to validate refresh tokens
	// The actual validation logic will be in the auth service since it needs database access
	return 0, fmt.Errorf("use auth service for refresh token validation")
}