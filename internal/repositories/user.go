package repositories

import (
	"payment-service/internal/database"
	"payment-service/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(tx *gorm.DB, user *models.User) error
	GetAll() ([]*models.User, error)
	GetByUserId(userId string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.GetDB()}
}

func (r *userRepository) Create(tx *gorm.DB, user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetByUserId(userId string) (*models.User, error) {
	var user *models.User
	if err := r.db.Where("user_id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
