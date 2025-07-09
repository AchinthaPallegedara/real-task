package models

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `json:"description"`
	Status      string    `gorm:"size:50;default:'pending'" json:"status"` // e.g., "pending", "in_progress", "completed"
	ProjectID   uint      `json:"project_id"`
	AssigneeID  *uint     `json:"assignee_id,omitempty"` // Pointer to allow NULL
	Assignee    *User     `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}