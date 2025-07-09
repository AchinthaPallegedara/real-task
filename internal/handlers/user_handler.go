package handlers

import (
	"net/http"

	"AchinthaPallegedara/real-task/internal/services"

	"github.com/gin-gonic/gin"
)

// GetUserProfile retrieves the authenticated user's profile
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id") // Get user ID from middleware context
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := services.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}