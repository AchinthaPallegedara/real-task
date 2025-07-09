package handlers

import (
	"AchinthaPallegedara/real-task/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateProject handles the POST /api/projects endpoint.
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