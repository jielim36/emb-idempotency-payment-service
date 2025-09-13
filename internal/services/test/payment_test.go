package services_test

import (
	"payment-service/internal/models"
	"payment-service/internal/repositories"
	mock_repositories "payment-service/internal/repositories/mock"
	"payment-service/internal/services"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func Initiate(
	t *testing.T,
	user *models.User,
	wallet *models.Wallet,
) (
	paymentRepo repositories.PaymentRepository,
	walletRepo repositories.WalletRepository,
	userRepo repositories.UserRepository,
	paymentService services.PaymentService,
) {
	paymentRepo = mock_repositories.NewMockPaymentRepository()
	walletRepo = mock_repositories.NewMockWalletRepository()
	userRepo = mock_repositories.NewMockUserRepository()
	paymentService = services.NewPaymentService(paymentRepo, walletRepo)

	// create testing data
	err := userRepo.Create(nil, user)
	assert.NoError(t, err, "failed to create user")
	err = walletRepo.Create(nil, wallet)
	assert.NoError(t, err, "failed to create wallet")

	return
}

func TestProcessPayment(t *testing.T) {
	ctx := &gin.Context{}

	user := &models.User{
		UserID: "user_1",
	}
	wallet := &models.Wallet{
		UserID:  user.UserID,
		Balance: decimal.NewFromInt(1000000),
	}
	req := &models.PaymentRequest{
		UserID:        user.UserID,
		Amount:        decimal.NewFromInt(100),
		TransactionID: "tx123",
	}

	_, _, _, paymentService := Initiate(t, user, wallet)

	// validate is new transaction id
	exist, err := paymentService.GetPaymentByTransactionID(req.TransactionID)
	assert.Nil(t, exist)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// First time: create new payment
	payment, err := paymentService.ProcessPayment(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, "tx123", payment.TransactionID)

	// Second time: should return existing payment (Idempotency)
	payment2, err := paymentService.ProcessPayment(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, payment2)
	assert.Equal(t, payment.UserID, payment2.UserID)
	assert.Equal(t, payment.TransactionID, payment2.TransactionID)
}

func TestProcessPaymentConcurrent(t *testing.T) {
	ctx := &gin.Context{}
	user := &models.User{
		UserID: "user_1",
	}
	wallet := &models.Wallet{
		UserID:  user.UserID,
		Balance: decimal.NewFromInt(1000000),
	}
	req := &models.PaymentRequest{
		UserID:        user.UserID,
		Amount:        decimal.NewFromInt(100),
		TransactionID: "tx123",
	}

	_, _, _, paymentService := Initiate(t, user, wallet)

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

			payment, err := paymentService.ProcessPayment(ctx, req)

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
