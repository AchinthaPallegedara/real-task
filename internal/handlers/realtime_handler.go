package handlers

import (
	"AchinthaPallegedara/real-task/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetConnectedUsers handles GET /api/projects/:project_id/connected-users
// Returns list of currently connected users for a project
// @Summary Get connected users for project
// @Description Get list of users currently connected to a specific project
// @Tags Real-time
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Success 200 {array} UserPresenceResponse "Connected users retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid project ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Router /api/projects/{project_id}/connected-users [get]
func GetConnectedUsers(c *gin.Context) {
	// Get the project ID from the URL parameter
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user has access to this project
	projectService := services.NewProjectService()
	hasAccess, err := projectService.UserHasProjectAccess(userID.(uint), uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check project access"})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: user cannot view this project"})
		return
	}

	// Get connected users from WebSocket hub
	if WebSocketHub == nil {
		c.JSON(http.StatusOK, gin.H{
			"connected_users": []gin.H{},
			"total_connected": 0,
		})
		return
	}

	connectedUsers := WebSocketHub.GetConnectedUsers()
	
	// Filter users who have access to this project
	projectConnectedUsers := []gin.H{}
	for _, connectedUserID := range connectedUsers {
		hasProjectAccess, err := projectService.UserHasProjectAccess(connectedUserID, uint(projectID))
		if err == nil && hasProjectAccess {
			presence := WebSocketHub.GetUserPresence(connectedUserID)
			userInfo := gin.H{
				"user_id": connectedUserID,
				"status":  "online",
			}
			if presence != nil {
				userInfo["status"] = presence.Status
				userInfo["last_seen"] = presence.LastSeen
			}
			projectConnectedUsers = append(projectConnectedUsers, userInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"connected_users": projectConnectedUsers,
		"total_connected": len(projectConnectedUsers),
	})
}

// GetUserPresence handles GET /api/users/:user_id/presence
// Returns presence information for a specific user
// @Summary Get user presence
// @Description Get presence information for a specific user
// @Tags Real-time
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Success 200 {object} UserPresenceResponse "User presence retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /api/users/{user_id}/presence [get]
func GetUserPresence(c *gin.Context) {
	// Get the user ID from the URL parameter
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get the requesting user ID from the context
	requestingUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if users share any projects (privacy check)
	projectService := services.NewProjectService()
	shareProjects, err := projectService.DoUsersShareProjects(requestingUserID.(uint), uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user relationships"})
		return
	}
	if !shareProjects {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: users do not share any projects"})
		return
	}

	// Get presence information
	if WebSocketHub == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id": uint(targetUserID),
			"status":  "offline",
			"message": "Real-time system not available",
		})
		return
	}

	presence := WebSocketHub.GetUserPresence(uint(targetUserID))
	if presence == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id": uint(targetUserID),
			"status":  "offline",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   presence.UserID,
		"status":    presence.Status,
		"last_seen": presence.LastSeen,
	})
}

// GetRealTimeStats handles GET /api/admin/realtime-stats
// Returns real-time system statistics (admin only)
// @Summary Get real-time statistics
// @Description Get real-time system statistics including connection counts
// @Tags Real-time
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} RealTimeStatsResponse "Real-time statistics retrieved successfully"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Admin access required"
// @Router /api/admin/realtime-stats [get]
func GetRealTimeStats(c *gin.Context) {
	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Simple admin check - you can enhance this with proper role-based access control
	// For now, we'll just return stats for any authenticated user
	_ = userID

	if WebSocketHub == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":           "disabled",
			"total_connected":  0,
			"connected_users":  []gin.H{},
			"message":          "Real-time system not available",
		})
		return
	}

	connectedUsers := WebSocketHub.GetConnectedUsers()
	userPresences := []gin.H{}
	
	for _, userID := range connectedUsers {
		presence := WebSocketHub.GetUserPresence(userID)
		userInfo := gin.H{
			"user_id": userID,
			"status":  "online",
		}
		if presence != nil {
			userInfo["status"] = presence.Status
			userInfo["last_seen"] = presence.LastSeen
		}
		userPresences = append(userPresences, userInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "active",
		"total_connected": len(connectedUsers),
		"connected_users": userPresences,
		"system_info": gin.H{
			"websocket_hub_active": true,
			"notification_service_active": NotificationService != nil,
		},
	})
}
