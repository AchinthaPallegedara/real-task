package services

import (
	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/models"
	"errors"

	"gorm.io/gorm"
)

// CollaborationService handles project collaboration logic
type CollaborationService struct {
	DB *gorm.DB
}

// NewCollaborationService creates a new CollaborationService instance
func NewCollaborationService() *CollaborationService {
	return &CollaborationService{
		DB: database.DB,
	}
}

// InviteUserInput defines the required fields for inviting a user to a project
type InviteUserInput struct {
	ProjectID uint   `json:"project_id" binding:"required"` // Project to invite user to
	UserEmail string `json:"user_email" binding:"required"` // Email of user to invite
	Message   string `json:"message,omitempty"`             // Optional invitation message
}

// InviteUser creates an invitation for a user to join a project
// Only project owners can send invitations
func (s *CollaborationService) InviteUser(input InviteUserInput, inviterID uint) (*models.ProjectInvitation, error) {
	// Verify that the inviter is the project owner
	var project models.Project
	result := s.DB.First(&project, input.ProjectID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, result.Error
	}

	if project.OwnerID != inviterID {
		return nil, errors.New("only project owners can invite collaborators")
	}

	// Find the user to invite by email
	var userToInvite models.User
	result = s.DB.Where("email = ?", input.UserEmail).First(&userToInvite)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found with that email")
		}
		return nil, result.Error
	}

	// Check if user is already the project owner
	if userToInvite.ID == project.OwnerID {
		return nil, errors.New("user is already the project owner")
	}

	// Check if user is already a collaborator
	var existingCollaborator models.ProjectCollaborator
	result = s.DB.Where("project_id = ? AND user_id = ?", input.ProjectID, userToInvite.ID).First(&existingCollaborator)
	if result.Error == nil {
		return nil, errors.New("user is already a collaborator on this project")
	}

	// Check if there's already a pending invitation
	var existingInvitation models.ProjectInvitation
	result = s.DB.Where("project_id = ? AND user_id = ? AND status = ?", input.ProjectID, userToInvite.ID, "pending").First(&existingInvitation)
	if result.Error == nil {
		return nil, errors.New("user already has a pending invitation for this project")
	}

	// Create the invitation
	invitation := models.ProjectInvitation{
		ProjectID: input.ProjectID,
		UserID:    userToInvite.ID,
		InviterID: inviterID,
		Status:    "pending",
		Message:   input.Message,
	}

	result = s.DB.Create(&invitation)
	if result.Error != nil {
		return nil, result.Error
	}

	// Load related data for response
	s.DB.Preload("Project").Preload("User").Preload("Inviter").First(&invitation, invitation.ID)

	return &invitation, nil
}

// GetPendingInvitations retrieves all pending invitations for a user
func (s *CollaborationService) GetPendingInvitations(userID uint) ([]models.ProjectInvitation, error) {
	var invitations []models.ProjectInvitation
	result := s.DB.Where("user_id = ? AND status = ?", userID, "pending").
		Preload("Project").Preload("Inviter").Find(&invitations)
	
	if result.Error != nil {
		return nil, result.Error
	}

	return invitations, nil
}

// RespondToInvitation allows a user to accept or decline a project invitation
func (s *CollaborationService) RespondToInvitation(invitationID uint, userID uint, accept bool) error {
	// Get the invitation
	var invitation models.ProjectInvitation
	result := s.DB.First(&invitation, invitationID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("invitation not found")
		}
		return result.Error
	}

	// Verify that the invitation is for this user
	if invitation.UserID != userID {
		return errors.New("this invitation is not for you")
	}

	// Verify that the invitation is still pending
	if invitation.Status != "pending" {
		return errors.New("invitation has already been responded to")
	}

	// Update invitation status
	status := "declined"
	if accept {
		status = "accepted"
	}

	result = s.DB.Model(&invitation).Update("status", status)
	if result.Error != nil {
		return result.Error
	}

	// If accepted, create a collaborator record
	if accept {
		collaborator := models.ProjectCollaborator{
			ProjectID: invitation.ProjectID,
			UserID:    invitation.UserID,
			Role:      "collaborator",
		}

		result = s.DB.Create(&collaborator)
		if result.Error != nil {
			// Rollback invitation status if collaborator creation fails
			s.DB.Model(&invitation).Update("status", "pending")
			return result.Error
		}
	}

	return nil
}

// GetProjectCollaborators retrieves all collaborators for a project
// Only project owners and collaborators can view the collaborator list
func (s *CollaborationService) GetProjectCollaborators(projectID uint, requestingUserID uint) ([]models.ProjectCollaborator, error) {
	// Check if the requesting user has access to the project
	hasAccess, err := s.checkProjectAccess(projectID, requestingUserID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("access denied: you don't have permission to view collaborators")
	}

	var collaborators []models.ProjectCollaborator
	result := s.DB.Where("project_id = ?", projectID).Preload("User").Find(&collaborators)
	if result.Error != nil {
		return nil, result.Error
	}

	return collaborators, nil
}

// RemoveCollaborator removes a user from a project
// Only project owners can remove collaborators
func (s *CollaborationService) RemoveCollaborator(projectID uint, collaboratorUserID uint, requestingUserID uint) error {
	// Verify that the requesting user is the project owner
	var project models.Project
	result := s.DB.First(&project, projectID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("project not found")
		}
		return result.Error
	}

	if project.OwnerID != requestingUserID {
		return errors.New("only project owners can remove collaborators")
	}

	// Cannot remove the project owner
	if collaboratorUserID == project.OwnerID {
		return errors.New("cannot remove the project owner")
	}

	// Remove the collaborator
	result = s.DB.Where("project_id = ? AND user_id = ?", projectID, collaboratorUserID).Delete(&models.ProjectCollaborator{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user is not a collaborator on this project")
	}

	return nil
}

// checkProjectAccess verifies if a user has access to a project (owner or collaborator)
func (s *CollaborationService) checkProjectAccess(projectID uint, userID uint) (bool, error) {
	var project models.Project
	result := s.DB.First(&project, projectID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, errors.New("project not found")
		}
		return false, result.Error
	}

	// Check if user is the project owner
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user is a collaborator
	var collaborator models.ProjectCollaborator
	result = s.DB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&collaborator)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil // User is not a collaborator
		}
		return false, result.Error
	}

	return true, nil // User is a collaborator
}
