package models

import (
	"gorm.io/gorm"
)

// Project represents a project that can have multiple tasks and collaborators
type Project struct {
	gorm.Model
	Name        string    `gorm:"size:255;not null" json:"name"`          // Project name
	Description string    `json:"description"`                            // Project description
	OwnerID     uint      `json:"owner_id"`                              // Foreign key to the project owner
	
	// Relationships
	Owner         User                  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`          // Belongs to a User (owner)
	Tasks         []Task               `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`        // Has many Tasks
	Collaborators []ProjectCollaborator `gorm:"foreignKey:ProjectID" json:"collaborators,omitempty"` // Has many Collaborators
}