package services_test

import (
	"payment-service/internal/models"
	mock_repositories "payment-service/internal/repositories/mock"
	"payment-service/internal/services"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestProcessPayment(t *testing.T) {
	repo := mock_repositories.NewMockPaymentRepository()
	service := services.NewPaymentService(repo)

	ctx := &gin.Context{}
	req := &models.PaymentRequest{
		UserID:        "user123",
		Amount:        decimal.NewFromInt(100),
		TransactionID: "tx123",
	}

	exist, err := service.GetPaymentByTransactionID(req.TransactionID)
	assert.Nil(t, exist)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// First time: create new payment
	payment, err := service.ProcessPayment(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, "tx123", payment.TransactionID)

	// Second time: should return existing payment (Idempotency)
	payment2, err := service.ProcessPayment(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, payment2)
	assert.Equal(t, payment.UserID, payment2.UserID)
	assert.Equal(t, payment.TransactionID, payment2.TransactionID)
}

func TestProcessPaymentConcurrent(t *testing.T) {
	repo := mock_repositories.NewMockPaymentRepository()
	service := services.NewPaymentService(repo)

	ctx := &gin.Context{}
	req := &models.PaymentRequest{
		UserID:        "user123",
		Amount:        decimal.NewFromInt(100),
		TransactionID: "tx_concurrent",
	}

	// We'll run 10 concurrent goroutines trying to process the same transaction
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []*models.Payment
	var errors []error

	goroutineCount := 10
	wg.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()

			payment, err := service.ProcessPayment(ctx, req)

			mu.Lock()
			results = append(results, payment)
			errors = append(errors, err)
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Verify: Only the first call should create a new payment, others should reuse it
	var nonNilPayments []*models.Payment
	for _, p := range results {
		if p != nil {
			nonNilPayments = append(nonNilPayments, p)
		}
	}

	assert.Equal(t, 10, len(results), "All goroutines should return a result")
	assert.Equal(t, 1, uniquePaymentsCount(nonNilPayments), "Only one unique payment should be created")

	// All errors should be nil
	for _, err := range errors {
		assert.NoError(t, err)
	}
}

// Helper: count unique payments by TransactionID
func uniquePaymentsCount(payments []*models.Payment) int {
	seen := make(map[string]bool)
	for _, p := range payments {
		if p != nil {
			seen[p.TransactionID] = true
		}
	}
	return len(seen)
}
