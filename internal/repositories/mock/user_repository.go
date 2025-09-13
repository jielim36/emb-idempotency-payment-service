package mock_repositories

import (
	"errors"
	"payment-service/internal/models"
	"payment-service/internal/repositories"
	"sync"

	"gorm.io/gorm"
)

type MockUserRepository struct {
	mu    sync.Mutex
	users map[string]*models.User // key: user.UserID
}

func NewMockUserRepository() repositories.UserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(tx *gorm.DB, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.UserID]; exists {
		return errors.New("user already exists")
	}

	user.ID = uint(len(m.users) + 1)
	m.users[user.UserID] = user
	return nil
}

func (m *MockUserRepository) GetAll() ([]*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var list []*models.User
	for _, user := range m.users {
		list = append(list, user)
	}
	return list, nil
}

func (m *MockUserRepository) GetByUserId(userId string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[userId]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}
