package handlers

import (
	"net/http"

	"AchinthaPallegedara/real-task/internal/services"
	"AchinthaPallegedara/real-task/internal/utils"

	"github.com/gin-gonic/gin"
)

// Request/Response structures
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TwoFactorToken string `json:"two_factor_token,omitempty"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ConfirmTwoFactorRequest struct {
	Token string `json:"token" binding:"required"`
}

type DisableTwoFactorRequest struct {
	Password string `json:"password" binding:"required"`
}

type VerifyTwoFactorRequest struct {
	Token string `json:"token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register handles user registration with email verification
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.RegisterUser(req.Email, req.Password)
	if err != nil {
		if err.Error() == "user already exists with this email" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists with this email"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email for verification link.",
		"user_id": user.ID,
		"email_verified": user.EmailVerified,
	})
}

// Login handles user login with 2FA support
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check if email is verified
	if !user.EmailVerified {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email not verified. Please check your email for verification link.",
			"email_verified": false,
		})
		return
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		if req.TwoFactorToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Two-factor authentication required",
				"two_factor_required": true,
			})
			return
		}

		// Verify 2FA token
		if err := services.VerifyTwoFactor(user.ID, req.TwoFactorToken); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid two-factor authentication token"})
			return
		}
	}

	tokenPair, err := services.GenerateTokenPair(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Set httpOnly cookies for tokens
	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		3600, // 1 hour
		"/",
		"",
		false, // Secure: false for local development
		true,  // HttpOnly: true
	)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		7*24*3600, // 7 days
		"/",
		"",
		false, // Secure: false for local development
		true,  // HttpOnly: true
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		// "token": tokenPair.AccessToken, // Keep this for backward compatibility
		"expires_at": tokenPair.ExpiresAt,
		"user": gin.H{
			"id": user.ID,
			"email": user.Email,
			"name": user.Name,
			"email_verified": user.EmailVerified,
			"two_factor_enabled": user.TwoFactorEnabled,
		},
	})
}

// VerifyEmail handles email verification
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
			return
		}
		token = req.Token
	}

	if err := services.VerifyEmail(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// ResendVerification resends email verification
func ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.ResendVerificationEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent successfully"})
}

// RequestPasswordReset initiates password reset process
func RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.RequestPasswordReset(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent if account exists"})
}

// ResetPassword handles password reset
func ResetPassword(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token = req.Token
		
		if err := services.ResetPassword(token, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		// GET request - show reset form (for web interface)
		c.JSON(http.StatusOK, gin.H{
			"message": "Password reset token is valid",
			"token": token,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// SetupTwoFactor initiates 2FA setup
func SetupTwoFactor(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	secret, qrCode, err := services.SetupTwoFactor(userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Two-factor authentication setup initiated",
		"secret": secret,
		"qr_code": qrCode,
		"instructions": "Scan the QR code with your authenticator app and enter the generated token to confirm setup",
	})
}

// ConfirmTwoFactor confirms 2FA setup
func ConfirmTwoFactor(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req ConfirmTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	backupCodes, err := services.ConfirmTwoFactor(userID.(uint), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Two-factor authentication enabled successfully",
		"backup_codes": backupCodes,
		"warning": "Save these backup codes in a secure location. Each code can only be used once.",
	})
}

// DisableTwoFactor disables 2FA
func DisableTwoFactor(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req DisableTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.DisableTwoFactor(userID.(uint), req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Two-factor authentication disabled successfully"})
}

// VerifyTwoFactor verifies 2FA token (for operations requiring 2FA)
func VerifyTwoFactor(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req VerifyTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.VerifyTwoFactor(userID.(uint), req.Token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Two-factor authentication verified"})
}

// GetAuthStatus returns current authentication status
func GetAuthStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := services.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id": user.ID,
			"email": user.Email,
			"name": user.Name,
			"email_verified": user.EmailVerified,
			"two_factor_enabled": user.TwoFactorEnabled,
			"is_active": user.IsActive,
			"last_login_at": user.LastLoginAt,
		},
	})
}

// ChangePassword allows authenticated users to change their password
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Verify current password
	if err := utils.CheckPasswordHash(req.CurrentPassword, user.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	user.Password = hashedPassword
	if err := services.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// RefreshToken handles refresh token requests
func RefreshToken(c *gin.Context) {
	// Try to get refresh token from cookie first, then from request body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Fallback to JSON request body
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
			return
		}
		refreshToken = req.RefreshToken
	}

	tokenPair, err := services.RefreshAccessToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new cookies
	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		3600, // 1 hour
		"/",
		"",
		false, // Secure: false for local development
		true,  // HttpOnly: true
	)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		7*24*3600, // 7 days
		"/",
		"",
		false, // Secure: false for local development
		true,  // HttpOnly: true
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"token": tokenPair.AccessToken, // Keep this for backward compatibility
		"expires_at": tokenPair.ExpiresAt,
	})
}

// Logout handles user logout (revokes refresh token)
func Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := services.RevokeRefreshToken(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Clear httpOnly cookies
	c.SetCookie(
		"access_token",
		"",
		-1, // Expire immediately
		"/",
		"",
		false, // Secure: false for local development
		true,  // HttpOnly: true
	)

	c.SetCookie(
		"refresh_token",
		"",
		-1, // Expire immediately
		"/",
		"",
		false, // Secure: false for local development
		true,  // HttpOnly: true
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// LogoutAllDevices handles logout from all devices (revokes all tokens)
func LogoutAllDevices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := services.RevokeAllUserTokens(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout from all devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices successfully"})
}