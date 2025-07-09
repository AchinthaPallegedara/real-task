package services

import (
	"AchinthaPallegedara/real-task/internal/database"
	"AchinthaPallegedara/real-task/internal/models"
)

func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if result := database.DB.First(&user, id); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser updates user information in the database
func UpdateUser(user *models.User) error {
	return database.DB.Save(user).Error
}