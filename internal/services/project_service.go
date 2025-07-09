package services

import (
	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/models"

	"gorm.io/gorm"
)

// ProjectService handles the business logic for projects.
type ProjectService struct {
	DB *gorm.DB
}

// NewProjectService creates a new ProjectService.
func NewProjectService() *ProjectService {
	return &ProjectService{
		DB: database.DB, // Using the global DB connection
	}
}

// CreateProjectInput defines the required fields for creating a project.
type CreateProjectInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateProject creates a new project for a given user.
func (s *ProjectService) CreateProject(input CreateProjectInput, ownerID uint) (*models.Project, error) {
	project := models.Project{
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     ownerID,
	}

	result := s.DB.Create(&project)
	if result.Error != nil {
		return nil, result.Error
	}

	return &project, nil
}

// GetProjectsForUser retrieves all projects owned by or collaborated on by a specific user
func (s *ProjectService) GetProjectsForUser(userID uint) ([]models.Project, error) {
	var ownedProjects []models.Project
	var collaboratedProjects []models.Project

    // Find all projects where the user is the owner
	result := s.DB.Where("owner_id = ?", userID).Preload("Owner").Find(&ownedProjects)
	if result.Error != nil {
		return nil, result.Error
	}

	// Find all projects where the user is a collaborator
	result = s.DB.Joins("JOIN project_collaborators ON projects.id = project_collaborators.project_id").
		Where("project_collaborators.user_id = ?", userID).
		Preload("Owner").
		Find(&collaboratedProjects)
	if result.Error != nil {
		return nil, result.Error
	}

	// Combine owned and collaborated projects
	allProjects := make([]models.Project, 0, len(ownedProjects)+len(collaboratedProjects))
	allProjects = append(allProjects, ownedProjects...)
	allProjects = append(allProjects, collaboratedProjects...)

	return allProjects, nil
}

// UserHasProjectAccess checks if a user has access to a specific project
func (s *ProjectService) UserHasProjectAccess(userID uint, projectID uint) (bool, error) {
	// Check if user is the owner
	var project models.Project
	result := s.DB.Where("id = ? AND owner_id = ?", projectID, userID).First(&project)
	if result.Error == nil {
		return true, nil // User is the owner
	}
	if result.Error != gorm.ErrRecordNotFound {
		return false, result.Error // Database error
	}

	// Check if user is a collaborator
	var collaborator models.ProjectCollaborator
	result = s.DB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&collaborator)
	if result.Error == nil {
		return true, nil // User is a collaborator
	}
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil // User has no access
	}

	return false, result.Error // Database error
}

// DoUsersShareProjects checks if two users share any projects
func (s *ProjectService) DoUsersShareProjects(userID1 uint, userID2 uint) (bool, error) {
	// Get projects for both users
	projects1, err := s.GetProjectsForUser(userID1)
	if err != nil {
		return false, err
	}

	projects2, err := s.GetProjectsForUser(userID2)
	if err != nil {
		return false, err
	}

	// Check for common projects
	projectSet := make(map[uint]bool)
	for _, project := range projects1 {
		projectSet[project.ID] = true
	}

	for _, project := range projects2 {
		if projectSet[project.ID] {
			return true, nil // Found a shared project
		}
	}

	return false, nil // No shared projects
}

// UpdateProjectInput defines the fields that can be updated for a project
type UpdateProjectInput struct {
	Name        *string `json:"name"`        // Optional: use pointer to detect if field was provided
	Description *string `json:"description"` // Optional: use pointer to detect if field was provided
}

// GetProjectByID retrieves a single project by ID with access control
func (s *ProjectService) GetProjectByID(projectID uint, userID uint) (*models.Project, error) {
	// First check if user has access to this project
	hasAccess, err := s.UserHasProjectAccess(userID, projectID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, gorm.ErrRecordNotFound // Return not found if no access
	}

	var project models.Project
	result := s.DB.Where("id = ?", projectID).Preload("Owner").First(&project)
	if result.Error != nil {
		return nil, result.Error
	}

	return &project, nil
}

// UpdateProject updates an existing project (only project owner can update)
func (s *ProjectService) UpdateProject(projectID uint, input UpdateProjectInput, userID uint) (*models.Project, error) {
	// Check if the user is the project owner (only owners can update projects)
	var project models.Project
	result := s.DB.Where("id = ? AND owner_id = ?", projectID, userID).First(&project)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound // Project not found or user is not the owner
		}
		return nil, result.Error
	}

	// Update only the fields that were provided
	updateData := make(map[string]interface{})
	if input.Name != nil {
		updateData["name"] = *input.Name
	}
	if input.Description != nil {
		updateData["description"] = *input.Description
	}

	// Perform the update
	if len(updateData) > 0 {
		result = s.DB.Model(&project).Updates(updateData)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	// Reload the project with owner information
	result = s.DB.Where("id = ?", projectID).Preload("Owner").First(&project)
	if result.Error != nil {
		return nil, result.Error
	}

	return &project, nil
}

// DeleteProject deletes a project (only project owner can delete)
func (s *ProjectService) DeleteProject(projectID uint, userID uint) error {
	// Check if the user is the project owner (only owners can delete projects)
	var project models.Project
	result := s.DB.Where("id = ? AND owner_id = ?", projectID, userID).First(&project)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound // Project not found or user is not the owner
		}
		return result.Error
	}

	// Delete related data first (CASCADE behavior)
	// Delete project collaborators
	s.DB.Where("project_id = ?", projectID).Delete(&models.ProjectCollaborator{})
	
	// Delete project invitations
	s.DB.Where("project_id = ?", projectID).Delete(&models.ProjectInvitation{})
	
	// Delete tasks in the project
	s.DB.Where("project_id = ?", projectID).Delete(&models.Task{})

	// Finally, delete the project
	result = s.DB.Delete(&project)
	if result.Error != nil {
		return result.Error
	}

	return nil
}