package models

import (
	"time"

	"gorm.io/gorm"
)

// ProjectCollaborator represents the many-to-many relationship between projects and users
// This allows multiple users to collaborate on the same project
type ProjectCollaborator struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProjectID uint           `gorm:"not null" json:"project_id"`         // Foreign key to Project
	UserID    uint           `gorm:"not null" json:"user_id"`            // Foreign key to User
	Role      string         `gorm:"size:50;default:'collaborator'" json:"role"` // Role: 'owner', 'collaborator', 'viewer'
	JoinedAt  time.Time      `gorm:"autoCreateTime" json:"joined_at"`    // When the user joined the project
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ProjectInvitation represents pending invitations to join a project
// This allows project owners to invite users who can accept or decline
type ProjectInvitation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProjectID uint           `gorm:"not null" json:"project_id"`              // Foreign key to Project
	UserID    uint           `gorm:"not null" json:"user_id"`                 // Foreign key to User being invited
	InviterID uint           `gorm:"not null" json:"inviter_id"`              // Foreign key to User who sent invitation
	Status    string         `gorm:"size:20;default:'pending'" json:"status"` // Status: 'pending', 'accepted', 'declined'
	Message   string         `gorm:"type:text" json:"message,omitempty"`      // Optional invitation message
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`       // User being invited
	Inviter User    `gorm:"foreignKey:InviterID" json:"inviter,omitempty"` // User who sent the invitation
}

// TableName returns the table name for ProjectCollaborator
func (ProjectCollaborator) TableName() string {
	return "project_collaborators"
}

// TableName returns the table name for ProjectInvitation
func (ProjectInvitation) TableName() string {
	return "project_invitations"
}
