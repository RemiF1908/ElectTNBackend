package services

import (
	"errors"
	"fmt"

	"elect-tn-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(email, name string, isAdmin bool, oauthToken string) (*models.User, error) {
	user := &models.User{
		Email:      email,
		Name:       name,
		IsAdmin:    isAdmin,
		OAuthToken: oauthToken,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByUUID(userUUID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (s *UserService) CreateOrUpdateUser(email, name, oauthToken string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new user
			newUser := &models.User{
				Email:      email,
				Name:       name,
				OAuthToken: oauthToken,
			}
			if err := s.db.Create(newUser).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			return newUser, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update existing user
	if err := s.db.Model(&user).Where("uuid = ?", user.UUID).Updates(map[string]interface{}{
		"name":         name,
		"o_auth_token": oauthToken,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Refresh the object
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to refresh user: %w", err)
	}

	return &user, nil
}

func (s *UserService) MakeUserAdmin(userUUID uuid.UUID) error {
	return s.db.Model(&models.User{}).Where("uuid = ?", userUUID).Update("is_admin", true).Error
}
