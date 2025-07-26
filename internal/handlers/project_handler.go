package handlers

import (
	"AchinthaPallegedara/real-task/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateProject handles the POST /api/projects endpoint.
// @Summary Create a new project
// @Description Create a new project for the authenticated user
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateProjectInput true "Project details"
// @Success 201 {object} ProjectResponse "Project created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/projects [post]
func CreateProject(c *gin.Context) {
	var input services.CreateProjectInput

	// Bind the incoming JSON to the input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get the user ID from the context (set by the JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new service instance
	projectService := services.NewProjectService()

	// Call the service to create the project
	project, err := projectService.CreateProject(input, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProjects handles the GET /api/projects endpoint.
// @Summary Get all projects
// @Description Get all projects for the authenticated user (owned + collaborated)
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ProjectResponse "Projects retrieved successfully"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/projects [get]
func GetProjects(c *gin.Context) {
	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new service instance
	projectService := services.NewProjectService()

	// Call the service to retrieve projects
	projects, err := projectService.GetProjectsForUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject handles the GET /api/projects/:project_id endpoint.
// @Summary Get a specific project
// @Description Get details of a specific project by ID
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Success 200 {object} ProjectResponse "Project retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid project ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 404 {object} ErrorResponse "Project not found"
// @Router /api/projects/{project_id} [get]
func GetProject(c *gin.Context) {
	// Get project ID from URL parameter
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Convert project ID to uint
	var projectIDUint uint
	if _, err := fmt.Sscanf(projectID, "%d", &projectIDUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new service instance
	projectService := services.NewProjectService()

	// Call the service to retrieve the project
	project, err := projectService.GetProjectByID(projectIDUint, userID.(uint))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject handles the PUT /api/projects/:project_id endpoint.
// @Summary Update a project
// @Description Update an existing project's details
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Param request body services.CreateProjectInput true "Updated project details"
// @Success 200 {object} ProjectResponse "Project updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 404 {object} ErrorResponse "Project not found"
// @Router /api/projects/{project_id} [put]
func UpdateProject(c *gin.Context) {
	// Get project ID from URL parameter
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Convert project ID to uint
	var projectIDUint uint
	if _, err := fmt.Sscanf(projectID, "%d", &projectIDUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var input services.UpdateProjectInput

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

	// Create a new service instance
	projectService := services.NewProjectService()

	// Call the service to update the project
	project, err := projectService.UpdateProject(projectIDUint, input, userID.(uint))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or you don't have permission to update it"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles the DELETE /api/projects/:project_id endpoint.
// @Summary Delete a project
// @Description Delete an existing project (owner only)
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Success 200 {object} SuccessResponse "Project deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid project ID"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Failure 403 {object} ErrorResponse "Access denied"
// @Failure 404 {object} ErrorResponse "Project not found"
// @Router /api/projects/{project_id} [delete]
func DeleteProject(c *gin.Context) {
	// Get project ID from URL parameter
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Convert project ID to uint
	var projectIDUint uint
	if _, err := fmt.Sscanf(projectID, "%d", &projectIDUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get the user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a new service instance
	projectService := services.NewProjectService()

	// Call the service to delete the project
	err := projectService.DeleteProject(projectIDUint, userID.(uint))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or you don't have permission to delete it"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}