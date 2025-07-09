package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringSlice is a custom type for handling JSON string arrays in PostgreSQL
type StringSlice []string

// Scan implements the Scanner interface for database/sql
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan into StringSlice")
	}
}

// Value implements the driver Valuer interface
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// User represents the user model in the database
type User struct {
	gorm.Model
	
	// Basic Info
	Email     string    `gorm:"unique;not null" json:"email"`            // User's email address
	Password  string    `gorm:"not null" json:"-"`                       // Password hash (hidden from JSON)
	Name      string    `gorm:"size:255" json:"name,omitempty"`          // User's display name (optional)
	
	// Email Verification
	EmailVerified       bool      `gorm:"default:false" json:"email_verified"`           // Email verification status
	EmailVerificationToken string `gorm:"size:255" json:"-"`                           // Email verification token
	EmailVerificationExpiry *time.Time `json:"-"`                                      // Email verification token expiry
	
	// Two-Factor Authentication
	TwoFactorEnabled    bool      `gorm:"default:false" json:"two_factor_enabled"`       // 2FA status
	TwoFactorSecret     string    `gorm:"size:255" json:"-"`                            // TOTP secret key
	TwoFactorBackupCodes StringSlice `gorm:"type:json" json:"-"`                          // Backup codes for 2FA
	
	// Password Reset
	PasswordResetToken  string    `gorm:"size:255" json:"-"`                            // Password reset token
	PasswordResetExpiry *time.Time `json:"-"`                                          // Password reset token expiry
	
	// Refresh Token for JWT
	RefreshToken        string    `gorm:"size:512" json:"-"`                            // Refresh token for JWT
	RefreshTokenExpiry  *time.Time `json:"-"`                                          // Refresh token expiry
	
	// Security & Activity
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`                     // Last login timestamp
	LoginAttempts       int       `gorm:"default:0" json:"-"`                          // Failed login attempts
	LockedUntil         *time.Time `json:"-"`                                          // Account lock expiry
	
	// Account Status
	IsActive            bool      `gorm:"default:true" json:"is_active"`               // Account active status
	
	// Relationships
	OwnedProjects   []Project             `gorm:"foreignKey:OwnerID" json:"owned_projects,omitempty"`   // Projects owned by this user
	Collaborations  []ProjectCollaborator `gorm:"foreignKey:UserID" json:"collaborations,omitempty"`   // Projects this user collaborates on
	AssignedTasks   []Task               `gorm:"foreignKey:AssigneeID" json:"assigned_tasks,omitempty"` // Tasks assigned to this user
}