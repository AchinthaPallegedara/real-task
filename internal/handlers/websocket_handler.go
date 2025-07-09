package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"AchinthaPallegedara/real-task/internal/models"
	"AchinthaPallegedara/real-task/internal/notifications"
	"AchinthaPallegedara/real-task/internal/services"
	"AchinthaPallegedara/real-task/internal/utils"
	wsocket "AchinthaPallegedara/real-task/internal/websocket"

	"github.com/gin-gonic/gin"
)

// WebSocketHub is the global hub instance
var WebSocketHub *wsocket.Hub

// NotificationService is the global notification service instance
var NotificationService *notifications.NotificationService

// InitWebSocketHub initializes the global WebSocket hub
func InitWebSocketHub() {
	WebSocketHub = wsocket.NewHub()
	go WebSocketHub.Run() // Start the hub in a goroutine
}

// InitNotificationService initializes the global notification service
func InitNotificationService() {
	NotificationService = notifications.NewNotificationService(3) // 3 worker goroutines
	NotificationService.Start()
}

// HandleWebSocket handles WebSocket connection requests
func HandleWebSocket(c *gin.Context) {
	// Get token from Authorization header (standard API approach)
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Use: Bearer <token>"})
		return
	}

	tokenString := parts[1]

	// Validate the token and extract user ID
	userID, err := utils.ValidateToken(tokenString)
	if err != nil {
		log.Printf("❌ Invalid WebSocket token for connection: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := wsocket.GetUpgrader().Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade connection to WebSocket for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}

	// Get user's accessible projects for real-time updates
	projectIDs, err := getUserAccessibleProjects(userID)
	if err != nil {
		log.Printf("❌ Failed to get accessible projects for user %d: %v", userID, err)
		conn.Close()
		return
	}

	// Create a new WebSocket client
	client := &wsocket.Client{
		ID:         strconv.Itoa(int(userID)),
		UserID:     userID,
		ProjectIDs: projectIDs,
		Conn:       conn,
		Send:       make(chan []byte, 256), // Buffered channel for outbound messages
		Hub:        WebSocketHub,
	}

	// Register client with the hub
	WebSocketHub.RegisterClient(client)

	// Start goroutines for reading and writing messages
	// These goroutines handle the actual WebSocket communication
	go client.StartReadPump()  // Read messages from client
	go client.StartWritePump() // Write messages to client

	log.Printf("✅ WebSocket connection established for user %d with access to %d projects", userID, len(projectIDs))
}

// getUserAccessibleProjects retrieves all project IDs that a user has access to
// This includes both owned projects and collaborated projects
func getUserAccessibleProjects(userID uint) ([]uint, error) {
	projectService := services.NewProjectService()
	projects, err := projectService.GetProjectsForUser(userID)
	if err != nil {
		return nil, err
	}

	// Extract project IDs
	projectIDs := make([]uint, len(projects))
	for i, project := range projects {
		projectIDs[i] = project.ID
	}

	return projectIDs, nil
}

// BroadcastTaskCreated sends a real-time notification when a task is created
func BroadcastTaskCreated(task *models.Task, creatorUserID uint) {
	if WebSocketHub != nil {
		taskData := map[string]interface{}{
			"task":       task,
			"created_by": creatorUserID,
			"message":    "A new task was created",
		}
		WebSocketHub.BroadcastTaskUpdate("task_created", task.ProjectID, creatorUserID, taskData)
	}
}

// BroadcastTaskUpdated sends a real-time notification when a task is updated
func BroadcastTaskUpdated(task *models.Task, updaterUserID uint, changes map[string]interface{}) {
	if WebSocketHub != nil {
		taskData := map[string]interface{}{
			"task":       task,
			"updated_by": updaterUserID,
			"changes":    changes,
			"message":    "A task was updated",
		}
		WebSocketHub.BroadcastTaskUpdate("task_updated", task.ProjectID, updaterUserID, taskData)
	}
	
	// 🔥 BACKGROUND NOTIFICATION: Send notification to assignee if someone else updated their task
	if NotificationService != nil && task.AssigneeID != nil && *task.AssigneeID != updaterUserID {
		NotificationService.SendTaskUpdatedNotification(*task.AssigneeID, updaterUserID, task.ProjectID, task.ID, task.Title, changes)
	}
	
	// 🔥 BACKGROUND NOTIFICATION: Send completion notification if task was marked as completed
	if statusChange, ok := changes["status"].(string); ok && statusChange == "completed" {
		if NotificationService != nil && task.AssigneeID != nil {
			NotificationService.SendTaskCompletedNotification(*task.AssigneeID, updaterUserID, task.ProjectID, task.ID, task.Title)
		}
	}
}

// BroadcastTaskDeleted sends a real-time notification when a task is deleted
func BroadcastTaskDeleted(projectID uint, taskID uint, deleterUserID uint) {
	if WebSocketHub != nil {
		taskData := map[string]interface{}{
			"task_id":    taskID,
			"project_id": projectID,
			"deleted_by": deleterUserID,
			"message":    "A task was deleted",
		}
		WebSocketHub.BroadcastTaskUpdate("task_deleted", projectID, deleterUserID, taskData)
	}
}

// BroadcastTaskAssigned sends a real-time notification when a task is assigned
func BroadcastTaskAssigned(task *models.Task, assignerUserID uint, assigneeUserID uint) {
	if WebSocketHub != nil {
		taskData := map[string]interface{}{
			"task":        task,
			"assigned_by": assignerUserID,
			"assigned_to": assigneeUserID,
			"message":     "A task was assigned to you",
		}
		WebSocketHub.BroadcastTaskUpdate("task_assigned", task.ProjectID, assignerUserID, taskData)
	}
	
	// 🔥 BACKGROUND NOTIFICATION: Send email notification to assignee
	if NotificationService != nil && assigneeUserID != assignerUserID {
		NotificationService.SendTaskAssignedNotification(assigneeUserID, assignerUserID, task.ProjectID, task.ID, task.Title)
	}
}

// BroadcastUserJoinedProject sends a real-time notification when a user joins a project
func BroadcastUserJoinedProject(projectID uint, userID uint, inviterUserID uint) {
	if WebSocketHub != nil {
		collaborationData := map[string]interface{}{
			"project_id":  projectID,
			"user_id":     userID,
			"invited_by":  inviterUserID,
			"message":     "A new collaborator joined the project",
		}
		WebSocketHub.BroadcastProjectUpdate("user_joined_project", projectID, inviterUserID, collaborationData)
	}
}

// BroadcastUserLeftProject sends a real-time notification when a user leaves a project
func BroadcastUserLeftProject(projectID uint, userID uint, removerUserID uint) {
	if WebSocketHub != nil {
		collaborationData := map[string]interface{}{
			"project_id": projectID,
			"user_id":    userID,
			"removed_by": removerUserID,
			"message":    "A collaborator left the project",
		}
		WebSocketHub.BroadcastProjectUpdate("user_left_project", projectID, removerUserID, collaborationData)
	}
}
