package handlers

import "time"

// Swagger response models for documentation
// These models are used only for generating Swagger documentation
// and don't include gorm.Model which causes issues with swag

// UserResponse represents a user in API responses
type UserResponse struct {
	ID              uint       `json:"id" example:"1"`
	CreatedAt       time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Email           string     `json:"email" example:"user@example.com"`
	Name            string     `json:"name" example:"John Doe"`
	EmailVerified   bool       `json:"email_verified" example:"true"`
	TwoFactorEnabled bool      `json:"two_factor_enabled" example:"false"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty" example:"2023-01-01T00:00:00Z"`
	IsActive        bool       `json:"is_active" example:"true"`
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID          uint         `json:"id" example:"1"`
	CreatedAt   time.Time    `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   time.Time    `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Name        string       `json:"name" example:"My Project"`
	Description string       `json:"description" example:"Project description"`
	OwnerID     uint         `json:"owner_id" example:"1"`
	Owner       *UserResponse `json:"owner,omitempty"`
}

// TaskResponse represents a task in API responses
type TaskResponse struct {
	ID          uint         `json:"id" example:"1"`
	CreatedAt   time.Time    `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   time.Time    `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Title       string       `json:"title" example:"Complete feature"`
	Description string       `json:"description" example:"Implement the new feature"`
	Status      string       `json:"status" example:"pending"`
	ProjectID   uint         `json:"project_id" example:"1"`
	AssigneeID  *uint        `json:"assignee_id,omitempty" example:"2"`
	Assignee    *UserResponse `json:"assignee,omitempty"`
}

// ProjectInvitationResponse represents a project invitation in API responses
type ProjectInvitationResponse struct {
	ID        uint         `json:"id" example:"1"`
	CreatedAt time.Time    `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time    `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	ProjectID uint         `json:"project_id" example:"1"`
	InviterID uint         `json:"inviter_id" example:"1"`
	InviteeID uint         `json:"invitee_id" example:"2"`
	Email     string       `json:"email" example:"invitee@example.com"`
	Role      string       `json:"role" example:"collaborator"`
	Status    string       `json:"status" example:"pending"`
	Project   *ProjectResponse `json:"project,omitempty"`
	Inviter   *UserResponse    `json:"inviter,omitempty"`
	Invitee   *UserResponse    `json:"invitee,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Message      string       `json:"message" example:"Login successful"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         UserResponse `json:"user"`
	ExpiresIn    int          `json:"expires_in" example:"3600"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid input"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// TwoFactorSetupResponse represents 2FA setup response
type TwoFactorSetupResponse struct {
	QRCodeURL   string   `json:"qr_code_url" example:"data:image/png;base64,iVBOR..."`
	SecretKey   string   `json:"secret_key" example:"JBSWY3DPEHPK3PXP"`
	BackupCodes []string `json:"backup_codes" example:"123456,789012,345678"`
}

// WebSocketMessage represents WebSocket message structure
type WebSocketMessage struct {
	Type      string      `json:"type" example:"task_update"`
	ProjectID uint        `json:"project_id" example:"1"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

// RealTimeStatsResponse represents real-time statistics
type RealTimeStatsResponse struct {
	TotalConnections    int            `json:"total_connections" example:"50"`
	ActiveProjects      int            `json:"active_projects" example:"10"`
	ConnectionsByProject map[string]int `json:"connections_by_project"`
}

// UserPresenceResponse represents user presence information
type UserPresenceResponse struct {
	UserID    uint      `json:"user_id" example:"1"`
	IsOnline  bool      `json:"is_online" example:"true"`
	LastSeen  time.Time `json:"last_seen" example:"2023-01-01T00:00:00Z"`
	ProjectID uint      `json:"project_id" example:"1"`
}

// OAuthProvidersResponse represents available OAuth providers
type OAuthProvidersResponse struct {
	Providers []OAuthProvider `json:"providers"`
}

// OAuthProvider represents an OAuth provider
type OAuthProvider struct {
	Name        string `json:"name" example:"google"`
	DisplayName string `json:"display_name" example:"Google"`
	Enabled     bool   `json:"enabled" example:"true"`
}

// OAuthURLResponse represents OAuth URL response
type OAuthURLResponse struct {
	AuthURL string `json:"auth_url" example:"https://accounts.google.com/oauth/authorize?..."`
	State   string `json:"state" example:"randomStateString"`
}
