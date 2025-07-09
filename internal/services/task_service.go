package services

import (
	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/models"
	"errors"

	"gorm.io/gorm"
)

// TaskService handles the business logic for tasks within projects
type TaskService struct {
	DB *gorm.DB
}

// NewTaskService creates a new TaskService instance
func NewTaskService() *TaskService {
	return &TaskService{
		DB: database.DB, // Using the global DB connection
	}
}

// CreateTaskInput defines the required fields for creating a task
type CreateTaskInput struct {
	Title       string `json:"title" binding:"required"`        // Task title is required
	Description string `json:"description"`                     // Optional description
	ProjectID   uint   `json:"-"`   // Must belong to a project
	AssigneeID  *uint  `json:"assignee_id,omitempty"`          // Optional assignee
}

// UpdateTaskInput defines the fields that can be updated for a task
type UpdateTaskInput struct {
	Title       *string `json:"title,omitempty"`        // Optional title update
	Description *string `json:"description,omitempty"`  // Optional description update
	Status      *string `json:"status,omitempty"`       // Optional status update (pending, in_progress, completed)
	AssigneeID  *uint   `json:"assignee_id,omitempty"`  // Optional assignee update
}

// CreateTask creates a new task within a project
// It verifies that the user has access to the project before creating the task
func (s *TaskService) CreateTask(input CreateTaskInput, userID uint) (*models.Task, error) {
	// First, verify that the user has access to the project (owner or collaborator)
	hasAccess, err := s.checkProjectAccess(input.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("access denied: user cannot create tasks in this project")
	}

	// If an assignee is specified, check if they have access to the project
	// If not, automatically send them an invitation
	if input.AssigneeID != nil {
		hasAccess, err := s.checkProjectAccess(input.ProjectID, *input.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			// User doesn't have access, so let's invite them automatically
			err := s.autoInviteUserToProject(input.ProjectID, *input.AssigneeID, userID)
			if err != nil {
				return nil, errors.New("failed to invite assignee to project: " + err.Error())
			}
			// Note: The invitation is pending, but we'll still create the task
			// The assignee will see the task once they accept the invitation
		}
	}

	// Create the task
	task := models.Task{
		Title:       input.Title,
		Description: input.Description,
		ProjectID:   input.ProjectID,
		AssigneeID:  input.AssigneeID,
		Status:      "pending", // Default status
	}

	result := s.DB.Create(&task)
	if result.Error != nil {
		return nil, result.Error
	}

	// Load the assignee information if present
	if task.AssigneeID != nil {
		s.DB.Preload("Assignee").First(&task, task.ID)
	}

	return &task, nil
}

// GetTasksForProject retrieves all tasks for a specific project
// Only users with access to the project can view its tasks
func (s *TaskService) GetTasksForProject(projectID uint, userID uint) ([]models.Task, error) {
	// Verify that the user has access to the project
	hasAccess, err := s.checkProjectAccess(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("access denied: user cannot view tasks in this project")
	}

	var tasks []models.Task
	// Preload assignee information for each task
	result := s.DB.Where("project_id = ?", projectID).Preload("Assignee").Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}

	return tasks, nil
}

// UpdateTask updates an existing task
// Only users with access to the project can update its tasks
func (s *TaskService) UpdateTask(taskID uint, input UpdateTaskInput, userID uint) (*models.Task, error) {
	// First, get the task to verify it exists and get its project ID
	var task models.Task
	result := s.DB.First(&task, taskID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, result.Error
	}

	// Verify that the user has access to the project
	hasAccess, err := s.checkProjectAccess(task.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("access denied: user cannot update tasks in this project")
	}

	// If assignee is being updated, check if they have access to the project
	// If not, automatically send them an invitation
	if input.AssigneeID != nil {
		hasAccess, err := s.checkProjectAccess(task.ProjectID, *input.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			// User doesn't have access, so let's invite them automatically
			err := s.autoInviteUserToProject(task.ProjectID, *input.AssigneeID, userID)
			if err != nil {
				return nil, errors.New("failed to invite assignee to project: " + err.Error())
			}
			// Note: The invitation is pending, but we'll still update the task assignment
			// The assignee will see the task once they accept the invitation
		}
	}

	// Update only the fields that are provided
	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Status != nil {
		// Validate status values
		validStatuses := map[string]bool{"pending": true, "in_progress": true, "completed": true}
		if !validStatuses[*input.Status] {
			return nil, errors.New("invalid status: must be 'pending', 'in_progress', or 'completed'")
		}
		updates["status"] = *input.Status
	}
	if input.AssigneeID != nil {
		updates["assignee_id"] = *input.AssigneeID
	}

	// Perform the update
	result = s.DB.Model(&task).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}

	// Reload the task with assignee information
	s.DB.Preload("Assignee").First(&task, taskID)

	return &task, nil
}

// DeleteTask deletes a task from a project
// Only users with access to the project can delete its tasks
func (s *TaskService) DeleteTask(taskID uint, userID uint) error {
	// First, get the task to verify it exists and get its project ID
	var task models.Task
	result := s.DB.First(&task, taskID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("task not found")
		}
		return result.Error
	}

	// Verify that the user has access to the project
	hasAccess, err := s.checkProjectAccess(task.ProjectID, userID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("access denied: user cannot delete tasks in this project")
	}

	// Delete the task
	result = s.DB.Delete(&task)
	return result.Error
}

// GetTaskByID retrieves a specific task by its ID
// Only users with access to the project can view the task
func (s *TaskService) GetTaskByID(taskID uint, userID uint) (*models.Task, error) {
	var task models.Task
	result := s.DB.Preload("Assignee").First(&task, taskID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, result.Error
	}

	// Verify that the user has access to the project
	hasAccess, err := s.checkProjectAccess(task.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("access denied: user cannot view this task")
	}

	return &task, nil
}

// checkProjectAccess is a helper function that verifies if a user has access to a project
// This includes both project owners and collaborators
func (s *TaskService) checkProjectAccess(projectID uint, userID uint) (bool, error) {
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

// autoInviteUserToProject automatically sends a project invitation to a user
// This is called when assigning a task to a user who doesn't have project access
func (s *TaskService) autoInviteUserToProject(projectID uint, userID uint, inviterID uint) error {
	// Check if user exists in the database
	var user models.User
	result := s.DB.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return result.Error
	}

	// Check if there's already a pending invitation
	var existingInvitation models.ProjectInvitation
	result = s.DB.Where("project_id = ? AND user_id = ? AND status = ?", projectID, userID, "pending").First(&existingInvitation)
	if result.Error == nil {
		// Invitation already exists, no need to create another one
		return nil
	}

	// Create the invitation with an automatic message
	invitation := models.ProjectInvitation{
		ProjectID: projectID,
		UserID:    userID,
		InviterID: inviterID,
		Status:    "pending",
		Message:   "You have been invited to collaborate on this project because a task was assigned to you.",
	}

	result = s.DB.Create(&invitation)
	return result.Error
}
