package repositories

import (
	"payment-service/internal/database"
	"payment-service/internal/models"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetAll() ([]*models.Payment, error)
	GetByTransactionID(transactionID string) (*models.Payment, error)
	GetByUserID(userID string) ([]models.Payment, error)
	Update(payment *models.Payment) error
	Delete(id uint) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository() PaymentRepository {
	return &paymentRepository{
		db: database.GetDB(),
	}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetAll() ([]*models.Payment, error) {
	var payments []*models.Payment
	if err := r.db.Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) GetByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByUserID(userID string) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.db.Where("user_id = ?", userID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Payment{}, id).Error
}
