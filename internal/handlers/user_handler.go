package handlers

import (
	"net/http"

	"AchinthaPallegedara/real-task/internal/services"

	"github.com/gin-gonic/gin"
)

// GetUserProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Description Get the profile information of the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /api/profile [get]
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