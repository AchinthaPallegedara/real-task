package main

import (
	"log"
	"os"

	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/handlers"
	"AchinthaPallegedara/real-task/internal/middleware"
	"AchinthaPallegedara/real-task/internal/models"
	"AchinthaPallegedara/real-task/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Set Gin mode based on environment
	if mode := os.Getenv("GIN_MODE"); mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		
	} else if mode == "" {
		gin.SetMode(gin.DebugMode) // Explicitly set debug mode for development
	}

	// Connect to database
	database.ConnectDB()

	// Auto-migrate database tables (for development, use proper migrations in production)
	err = database.DB.AutoMigrate(
		&models.User{}, 
		&models.Project{}, 
		&models.Task{},
		&models.ProjectCollaborator{},
		&models.ProjectInvitation{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	
	log.Println("🔄 Database migration completed with refresh token support!")

	// Initialize Gin router
	router := gin.New()
	
	// Add middleware manually for better control
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware()) // Add CORS middleware
	
	// Configure trusted proxies for security
	router.SetTrustedProxies([]string{"127.0.0.1", "::1"}) // Only trust localhost

	// 🔥 Initialize WebSocket Hub for real-time features
	handlers.InitWebSocketHub()
	log.Println("🚀 Real-time WebSocket system initialized!")

	// 🔥 Initialize Notification Service for background notifications
	handlers.InitNotificationService()
	log.Println("📧 Background notification system initialized!")

	// 🔥 Initialize OAuth configurations
	services.InitOAuthConfigs()
	log.Println("🔐 OAuth providers initialized!")

	// Public routes
	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)
	
	// Public auth routes (no authentication required)
	auth := router.Group("/api/auth")
	{
		auth.GET("/verify-email", handlers.VerifyEmail)                    // Email verification (GET)
		auth.POST("/verify-email", handlers.VerifyEmail)                   // Email verification (POST)
		auth.POST("/resend-verification", handlers.ResendVerification)     // Resend verification email
		auth.POST("/request-password-reset", handlers.RequestPasswordReset) // Request password reset
		auth.GET("/reset-password", handlers.ResetPassword)                // Password reset form (GET)
		auth.POST("/reset-password", handlers.ResetPassword)               // Password reset (POST)
		auth.POST("/refresh", handlers.RefreshToken)                       // Refresh access token
		
		// 🔥 OAuth routes
		auth.GET("/providers", handlers.OAuthProviders)                    // Get available OAuth providers
		auth.GET("/google/url", handlers.GoogleAuthURL)                    // Get Google OAuth URL
		auth.GET("/google/callback", handlers.GoogleAuthCallback)          // Google OAuth callback
		auth.POST("/google", handlers.GoogleAuthLogin)                     // Google OAuth login (POST)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		// User routes
		protected.GET("/profile", handlers.GetUserProfile)
		protected.GET("/auth/status", handlers.GetAuthStatus)           // Get authentication status
		protected.POST("/auth/change-password", handlers.ChangePassword) // Change password
		protected.POST("/auth/logout", handlers.Logout)                // Logout (revoke refresh token)
		protected.POST("/auth/logout-all", handlers.LogoutAllDevices)  // Logout from all devices
		
		// Two-Factor Authentication routes
		protected.POST("/auth/2fa/setup", handlers.SetupTwoFactor)      // Setup 2FA
		protected.POST("/auth/2fa/confirm", handlers.ConfirmTwoFactor)  // Confirm 2FA setup
		protected.POST("/auth/2fa/disable", handlers.DisableTwoFactor)  // Disable 2FA
		protected.POST("/auth/2fa/verify", handlers.VerifyTwoFactor)    // Verify 2FA token
		
		// Project routes
		protected.POST("/projects", handlers.CreateProject)            // Create new project
        protected.GET("/projects", handlers.GetProjects)              // Get all user's projects
		protected.GET("/projects/:project_id", handlers.GetProject)   // Get specific project
		protected.PUT("/projects/:project_id", handlers.UpdateProject) // Update project
		protected.DELETE("/projects/:project_id", handlers.DeleteProject) // Delete project
		
		// Task routes
		protected.POST("/projects/:project_id/tasks", handlers.CreateTask)         // Create task in project
		protected.GET("/projects/:project_id/tasks", handlers.GetTasksForProject) // Get all tasks in project
		protected.GET("/tasks/:task_id", handlers.GetTask)                        // Get specific task
		protected.PUT("/tasks/:task_id", handlers.UpdateTask)                     // Update task
		protected.DELETE("/tasks/:task_id", handlers.DeleteTask)                  // Delete task
		
		// Collaboration routes
		protected.POST("/projects/:project_id/invite", handlers.InviteUserToProject)               // Invite user to project
		protected.GET("/invitations", handlers.GetPendingInvitations)                              // Get user's pending invitations
		protected.POST("/invitations/:invitation_id/respond", handlers.RespondToInvitation)        // Accept/decline invitation
		protected.GET("/projects/:project_id/collaborators", handlers.GetProjectCollaborators)    // Get project collaborators
		protected.DELETE("/projects/:project_id/collaborators/:user_id", handlers.RemoveCollaborator) // Remove collaborator
		
		// 🔥 Real-time features
		protected.GET("/projects/:project_id/connected-users", handlers.GetConnectedUsers) // Get connected users for project
		protected.GET("/users/:user_id/presence", handlers.GetUserPresence) // Get user presence
		protected.GET("/admin/realtime-stats", handlers.GetRealTimeStats) // Real-time system stats
	}

	// WebSocket endpoint - handled separately to support both header and query parameter authentication
	router.GET("/api/ws", handlers.HandleWebSocket) // WebSocket connection endpoint

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}