package middleware

import (
	"net/http"
	"strings"

	"AchinthaPallegedara/real-task/internal/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates the JWT token from the Authorization header or cookies
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If no token in header, try to get from cookie
		if tokenString == "" {
			cookie, err := c.Cookie("access_token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}
			tokenString = cookie
		}

		userID, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID) // Store user ID in context for later use
		c.Next()
	}
}