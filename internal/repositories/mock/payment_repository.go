package mock_repositories

import (
	"errors"
	"payment-service/internal/models"
	"payment-service/internal/repositories"

	"gorm.io/gorm"
)

type MockPaymentRepository struct {
	payments map[string]*models.Payment
}

func NewMockPaymentRepository() repositories.PaymentRepository {
	return &MockPaymentRepository{payments: make(map[string]*models.Payment)}
}

func (f *MockPaymentRepository) Create(payment *models.Payment) error {
	if _, exists := f.payments[payment.TransactionID]; exists {
		return errors.New("duplicate transaction ID")
	}
	f.payments[payment.TransactionID] = payment
	return nil
}

func (f *MockPaymentRepository) GetAll() ([]*models.Payment, error) {
	var result []*models.Payment
	for _, p := range f.payments {
		result = append(result, p)
	}
	return result, nil
}

func (f *MockPaymentRepository) GetByTransactionID(transactionID string) (*models.Payment, error) {
	if p, exists := f.payments[transactionID]; exists {
		return p, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *MockPaymentRepository) Update(tx *gorm.DB, payment *models.Payment) error {
	if _, exists := f.payments[payment.TransactionID]; !exists {
		return gorm.ErrRecordNotFound
	}
	f.payments[payment.TransactionID] = payment
	return nil
}

func (f *MockPaymentRepository) Delete(id uint) error {
	for k, v := range f.payments {
		if v.ID == id {
			delete(f.payments, k)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}
