package repositories

import (
	"payment-service/internal/database"
	"payment-service/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository interface {
	Create(tx *gorm.DB, wallet *models.Wallet) error
	GetForUpdate(tx *gorm.DB, userID string) (*models.Wallet, error)
	GetByUserId(userId string) (*models.Wallet, error)
	UpdateBalance(tx *gorm.DB, wallet *models.Wallet) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository() WalletRepository {
	return &walletRepository{db: database.GetDB()}
}

func (r *walletRepository) Create(tx *gorm.DB, wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

// GetForUpdate
// lock the selected rows for the duration of the transaction.
// This can be used in scenarios where you are preparing to update the rows and want to prevent other transactions from modifying them until your transaction is complete.
func (r *walletRepository) GetForUpdate(tx *gorm.DB, userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserId(userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	if err := r.db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(tx *gorm.DB, wallet *models.Wallet) error {
	return tx.Model(wallet).Updates(map[string]interface{}{
		"balance": wallet.Balance,
	}).Error
}
