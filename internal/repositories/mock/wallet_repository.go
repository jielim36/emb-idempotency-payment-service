package mock_repositories

import (
	"errors"
	"payment-service/internal/models"
	"payment-service/internal/repositories"
	"sync"

	"gorm.io/gorm"
)

type MockWalletRepository struct {
	mu      sync.Mutex
	wallets map[string]*models.Wallet
}

func NewMockWalletRepository() repositories.WalletRepository {
	return &MockWalletRepository{
		wallets: make(map[string]*models.Wallet),
	}
}

func (m *MockWalletRepository) GetForUpdate(tx *gorm.DB, userID string) (*models.Wallet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if wallet, ok := m.wallets[userID]; ok {
		copyWallet := *wallet
		return &copyWallet, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockWalletRepository) GetByUserId(userID string) (*models.Wallet, error) {
	if wallet, ok := m.wallets[userID]; ok {
		copyWallet := *wallet
		return &copyWallet, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockWalletRepository) UpdateBalance(tx *gorm.DB, wallet *models.Wallet) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.wallets[wallet.UserID]; !ok {
		return errors.New("wallet not found")
	}

	m.wallets[wallet.UserID].Balance = wallet.Balance
	return nil
}

func (m *MockWalletRepository) Create(tx *gorm.DB, wallet *models.Wallet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.wallets[wallet.UserID] = wallet

	return nil
}
