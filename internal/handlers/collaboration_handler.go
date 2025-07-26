package handlers

import (
	"AchinthaPallegedara/real-task/internal/models"
	"AchinthaPallegedara/real-task/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InviteUserToProject handles the POST /api/projects/:project_id/invite endpoint
// Allows project owners to invite users to collaborate on a project
// @Summary Invite user to project
// @Description Invite a user to collaborate on a project (project owner only)
// @Tags Collaboration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Param request body services.InviteUserInput true "Invitation details"
// @Success 201 {object} ProjectInvitationResponse "Invitation sent successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/projects/{project_id}/invite [post]
func InviteUserToProject(c *gin.Context) {
	// Get the project ID from the URL parameter
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var input services.InviteUserInput
	// Bind the incoming JSON to the input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Set the project ID from the URL parameter
	input.ProjectID = uint(projectID)

	// Get the user ID from the context (the inviter)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new collaboration service instance
	collaborationService := services.NewCollaborationService()

	// Call the service to create the invitation
	invitation, err := collaborationService.InviteUser(input, userID.(uint))
	if err != nil {
		// Check for specific error types to return appropriate HTTP status codes
		if err.Error() == "only project owners can invite collaborators" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "project not found" || err.Error() == "user not found with that email" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user is already the project owner" ||
			err.Error() == "user is already a collaborator on this project" ||
			err.Error() == "user already has a pending invitation for this project" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send invitation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Invitation sent successfully",
		"invitation": invitation,
	})
}

// GetPendingInvitations handles the GET /api/invitations endpoint
// Retrieves all pending invitations for the authenticated user
func GetPendingInvitations(c *gin.Context) {
	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new collaboration service instance
	collaborationService := services.NewCollaborationService()

	// Call the service to retrieve pending invitations
	invitations, err := collaborationService.GetPendingInvitations(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invitations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
	})
}

// RespondToInvitation handles the POST /api/invitations/:invitation_id/respond endpoint
// Allows users to accept or decline project invitations
func RespondToInvitation(c *gin.Context) {
	// Get the invitation ID from the URL parameter
	invitationIDStr := c.Param("invitation_id")
	invitationID, err := strconv.ParseUint(invitationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	// Define the input structure for the response
	var input struct {
		Accept bool `json:"accept" binding:"required"` // true to accept, false to decline
	}

	// Bind the incoming JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new collaboration service instance
	collaborationService := services.NewCollaborationService()

	// Call the service to respond to the invitation
	err = collaborationService.RespondToInvitation(uint(invitationID), userID.(uint), input.Accept)
	if err != nil {
		// Check for specific error types
		if err.Error() == "invitation not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "this invitation is not for you" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invitation has already been responded to" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to respond to invitation"})
		return
	}

	// Return appropriate success message
	var message string
	if input.Accept {
		message = "Invitation accepted successfully. You are now a collaborator on this project."
		
		// 🔥 REAL-TIME FEATURE: Broadcast that user joined the project
		// We need to get the project ID from the invitation
		var invitation models.ProjectInvitation
		collaborationService.DB.First(&invitation, invitationID)
		BroadcastUserJoinedProject(invitation.ProjectID, userID.(uint), invitation.InviterID)
	} else {
		message = "Invitation declined successfully."
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

// GetProjectCollaborators handles the GET /api/projects/:project_id/collaborators endpoint
// Retrieves all collaborators for a specific project
func GetProjectCollaborators(c *gin.Context) {
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

	// Create a new collaboration service instance
	collaborationService := services.NewCollaborationService()

	// Call the service to retrieve project collaborators
	collaborators, err := collaborationService.GetProjectCollaborators(uint(projectID), userID.(uint))
	if err != nil {
		// Check for specific error types
		if err.Error() == "access denied: you don't have permission to view collaborators" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve collaborators"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collaborators": collaborators,
	})
}

// RemoveCollaborator handles the DELETE /api/projects/:project_id/collaborators/:user_id endpoint
// Allows project owners to remove collaborators from a project
func RemoveCollaborator(c *gin.Context) {
	// Get the project ID from the URL parameter
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get the collaborator user ID from the URL parameter
	collaboratorUserIDStr := c.Param("user_id")
	collaboratorUserID, err := strconv.ParseUint(collaboratorUserIDStr, 10, 32)
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

	// Create a new collaboration service instance
	collaborationService := services.NewCollaborationService()

	// Call the service to remove the collaborator
	err = collaborationService.RemoveCollaborator(uint(projectID), uint(collaboratorUserID), requestingUserID.(uint))
	if err != nil {
		// Check for specific error types
		if err.Error() == "only project owners can remove collaborators" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot remove the project owner" ||
			err.Error() == "user is not a collaborator on this project" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove collaborator"})
		return
	}

	// 🔥 REAL-TIME FEATURE: Broadcast that user left the project
	BroadcastUserLeftProject(uint(projectID), uint(collaboratorUserID), requestingUserID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Collaborator removed successfully"})
}
