package handlers

import (
	"AchinthaPallegedara/real-task/internal/models"
	"AchinthaPallegedara/real-task/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateTask handles the POST /api/projects/:project_id/tasks endpoint
// Creates a new task within a specific project
// @Summary Create a new task
// @Description Create a new task within a specific project
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Param request body services.CreateTaskInput true "Task details"
// @Success 201 {object} TaskResponse "Task created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/projects/{project_id}/tasks [post]
func CreateTask(c *gin.Context) {
	// Get the project ID from the URL parameter
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var input services.CreateTaskInput
	// Bind the incoming JSON to the input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Set the project ID from the URL parameter
	input.ProjectID = uint(projectID)

	// Get the user ID from the context (set by the JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new task service instance
	taskService := services.NewTaskService()

	// Call the service to create the task
	task, err := taskService.CreateTask(input, userID.(uint))
	if err != nil {
		// Check for specific error types to return appropriate HTTP status codes
		if err.Error() == "access denied: user cannot create tasks in this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assignee user not found"})
			return
		}
		// Check if it's an auto-invitation error
		if len(err.Error()) > 40 && err.Error()[:40] == "failed to invite assignee to project: " {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Task created but failed to send invitation to assignee: " + err.Error()[40:],
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Check if the task has an assignee who might have been auto-invited
	response := gin.H{"task": task}
	if task.AssigneeID != nil {
		// Check if there's a pending invitation for the assignee
		var invitation models.ProjectInvitation
		taskService := services.NewTaskService()
		result := taskService.DB.Where("project_id = ? AND user_id = ? AND status = ?", task.ProjectID, *task.AssigneeID, "pending").First(&invitation)
		if result.Error == nil {
			response["message"] = "Task created successfully. An invitation has been sent to the assignee to join this project."
		}
	}

	// 🔥 REAL-TIME FEATURE: Broadcast task creation to all project collaborators
	BroadcastTaskCreated(task, userID.(uint))

	c.JSON(http.StatusCreated, response)
}

// GetTasksForProject handles the GET /api/projects/:project_id/tasks endpoint
// Retrieves all tasks for a specific project
// @Summary Get all tasks for a project
// @Description Get all tasks within a specific project
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Success 200 {array} TaskResponse "Tasks retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid project ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/projects/{project_id}/tasks [get]
func GetTasksForProject(c *gin.Context) {
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

	// Create a new task service instance
	taskService := services.NewTaskService()

	// Call the service to retrieve tasks for the project
	tasks, err := taskService.GetTasksForProject(uint(projectID), userID.(uint))
	if err != nil {
		// Check for specific error types
		if err.Error() == "access denied: user cannot view tasks in this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask handles the GET /api/tasks/:task_id endpoint
// Retrieves a specific task by its ID
// @Summary Get a specific task
// @Description Get details of a specific task by ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id path int true "Task ID"
// @Success 200 {object} TaskResponse "Task retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid task ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 404 {object} ErrorResponse "Task not found"
// @Router /api/tasks/{task_id} [get]
func GetTask(c *gin.Context) {
	// Get the task ID from the URL parameter
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new task service instance
	taskService := services.NewTaskService()

	// Call the service to retrieve the task
	task, err := taskService.GetTaskByID(uint(taskID), userID.(uint))
	if err != nil {
		// Check for specific error types
		if err.Error() == "access denied: user cannot view this task" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask handles the PUT /api/tasks/:task_id endpoint
// Updates an existing task
// @Summary Update a task
// @Description Update an existing task's details
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id path int true "Task ID"
// @Param request body services.UpdateTaskInput true "Updated task details"
// @Success 200 {object} TaskResponse "Task updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 404 {object} ErrorResponse "Task not found"
// @Router /api/tasks/{task_id} [put]
func UpdateTask(c *gin.Context) {
	// Get the task ID from the URL parameter
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var input services.UpdateTaskInput
	// Bind the incoming JSON to the input struct
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

	// Create a new task service instance
	taskService := services.NewTaskService()

	// Call the service to update the task
	task, err := taskService.UpdateTask(uint(taskID), input, userID.(uint))
	if err != nil {
		// Check for specific error types
		if err.Error() == "access denied: user cannot update tasks in this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assignee user not found"})
			return
		}
		if err.Error() == "invalid status: must be 'pending', 'in_progress', or 'completed'" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Check if it's an auto-invitation error
		if len(err.Error()) > 40 && err.Error()[:40] == "failed to invite assignee to project: " {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Task updated but failed to send invitation to assignee: " + err.Error()[40:],
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Check if the task assignment changed and an invitation might have been sent
	response := gin.H{"task": task}
	if input.AssigneeID != nil && task.AssigneeID != nil {
		// Check if there's a pending invitation for the assignee
		var invitation models.ProjectInvitation
		taskService := services.NewTaskService()
		result := taskService.DB.Where("project_id = ? AND user_id = ? AND status = ?", task.ProjectID, *task.AssigneeID, "pending").First(&invitation)
		if result.Error == nil {
			response["message"] = "Task updated successfully. An invitation has been sent to the new assignee to join this project."
		}
	}

	// 🔥 REAL-TIME FEATURE: Broadcast task update to all project collaborators
	// Create a map of what changed for the broadcast
	changes := make(map[string]interface{})
	if input.Title != nil {
		changes["title"] = *input.Title
	}
	if input.Description != nil {
		changes["description"] = *input.Description
	}
	if input.Status != nil {
		changes["status"] = *input.Status
	}
	if input.AssigneeID != nil {
		changes["assignee_id"] = *input.AssigneeID
		// If task was assigned to someone, send special assignment notification
		BroadcastTaskAssigned(task, userID.(uint), *input.AssigneeID)
	}
	
	BroadcastTaskUpdated(task, userID.(uint), changes)

	c.JSON(http.StatusOK, response)
}

// DeleteTask handles the DELETE /api/tasks/:task_id endpoint
// Deletes an existing task
// @Summary Delete a task
// @Description Delete an existing task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id path int true "Task ID"
// @Success 200 {object} SuccessResponse "Task deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid task ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 404 {object} ErrorResponse "Task not found"
// @Router /api/tasks/{task_id} [delete]
func DeleteTask(c *gin.Context) {
	// Get the task ID from the URL parameter
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new task service instance
	taskService := services.NewTaskService()

	// Call the service to delete the task
	err = taskService.DeleteTask(uint(taskID), userID.(uint))
	if err != nil {
		// Check for specific error types
		if err.Error() == "access denied: user cannot delete tasks in this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	// 🔥 REAL-TIME FEATURE: Broadcast task deletion to all project collaborators
	// We need to get the project ID before deletion for broadcasting
	var task models.Task
	taskService.DB.First(&task, taskID) // Get task info for project ID
	BroadcastTaskDeleted(task.ProjectID, uint(taskID), userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
