package services_test

import (
	"context"
	"fmt"
	"payment-service/internal/database"
	"payment-service/internal/models"
	"payment-service/internal/repositories"
	"payment-service/internal/services"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"gorm.io/gorm"
)

type TestContext struct {
	Ctx                  *gin.Context
	Container            testcontainers.Container
	EstimatedProcessTime time.Duration

	// Dependencies
	PaymentRepo    repositories.PaymentRepository
	WalletRepo     repositories.WalletRepository
	UserRepo       repositories.UserRepository
	PaymentService services.PaymentService
}

func Initiate(
	t *testing.T,
	user *models.User,
	wallet *models.Wallet,
) *TestContext {
	ctx, _ := gin.CreateTestContext(nil)

	db, container, err := database.InitTestDatabase()
	if err != nil {
		t.Fatalf("failed to init test DB: %v", err)
	}

	paymentRepo := repositories.NewPaymentRepository(db)
	walletRepo := repositories.NewWalletRepository(db)
	userRepo := repositories.NewUserRepository(db)
	paymentService := services.NewPaymentService(db, paymentRepo, walletRepo)

	// 初始化测试数据
	err = userRepo.Create(db, user)
	assert.NoError(t, err, "failed to create user")
	err = walletRepo.Create(db, wallet)
	assert.NoError(t, err, "failed to create wallet")

	return &TestContext{
		Ctx:                  ctx,
		Container:            container,
		EstimatedProcessTime: 4 * time.Second,

		// Dependencies
		PaymentRepo:    paymentRepo,
		WalletRepo:     walletRepo,
		UserRepo:       userRepo,
		PaymentService: paymentService,
	}
}

func TestFullFlow(t *testing.T) {
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

	tc := Initiate(t, user, wallet)
	defer func() {
		_ = tc.Container.Terminate(context.Background())
	}()

	t.Run("Validate new transaction ID should not exist", func(t *testing.T) {
		exist, err := tc.PaymentService.GetPaymentByTransactionID(req.TransactionID)
		assert.Nil(t, exist)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("First payment creation should succeed", func(t *testing.T) {
		payment, err := tc.PaymentService.ProcessPayment(tc.Ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, "tx123", payment.TransactionID)
	})
	time.Sleep(tc.EstimatedProcessTime)

	t.Run("Check payment process result and wallet balance", func(t *testing.T) {
		latestPayment, err := tc.PaymentService.GetPaymentByTransactionID(req.TransactionID)
		assert.NoError(t, err)

		latestWallet, err := tc.WalletRepo.GetByUserId(user.UserID)
		assert.NoError(t, err)

		expectedStatus := []models.PaymentStatus{models.StatusCompleted, models.StatusFailed}
		assert.True(t, slices.Contains(expectedStatus, latestPayment.Status), "Payment should be completed or failed")

		expectedBalance := wallet.Balance.Sub(req.Amount)
		assert.True(t, latestWallet.Balance.Equal(expectedBalance))
	})
}

func TestMakePaymentWithDuplicatedTransactionId(t *testing.T) {
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

	tc := Initiate(t, user, wallet)
	defer func() {
		_ = tc.Container.Terminate(context.Background())
	}()

	t.Run("Validate new transaction ID should not exist", func(t *testing.T) {
		exist, err := tc.PaymentService.GetPaymentByTransactionID(req.TransactionID)
		assert.Nil(t, exist)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	var payment *models.Payment
	t.Run("First time: create new payment", func(t *testing.T) {
		var err error
		payment, err = tc.PaymentService.ProcessPayment(tc.Ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, "tx123", payment.TransactionID)
	})

	t.Run("Second time: should return existing payment (Idempotency)", func(t *testing.T) {
		payment2, err := tc.PaymentService.ProcessPayment(tc.Ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, payment2)
		assert.Equal(t, payment.UserID, payment2.UserID)
		assert.Equal(t, payment.TransactionID, payment2.TransactionID)
	})

	time.Sleep(tc.EstimatedProcessTime)
	t.Run("Validate wallet balance only deducted once", func(t *testing.T) {
		latestWallet, err := tc.WalletRepo.GetByUserId(user.UserID)
		assert.NoError(t, err)

		expectedBalance := wallet.Balance.Sub(req.Amount)
		assert.True(t, latestWallet.Balance.Equal(expectedBalance), "wallet balance should only be deducted once")
	})
}

func TestProcessPaymentConcurrent(t *testing.T) {
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

	tc := Initiate(t, user, wallet)
	defer func() {
		_ = tc.Container.Terminate(context.Background())
	}()

	goroutineCount := 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []*models.Payment
	var errors []error

	t.Run("Run concurrent ProcessPayment calls", func(t *testing.T) {
		wg.Add(goroutineCount)

		for i := 0; i < goroutineCount; i++ {
			go func() {
				defer wg.Done()

				payment, err := tc.PaymentService.ProcessPayment(tc.Ctx, req)

				mu.Lock()
				results = append(results, payment)
				errors = append(errors, err)
				mu.Unlock()
			}()
		}

		wg.Wait()
	})

	time.Sleep(tc.EstimatedProcessTime)
	t.Run("Verify only one unique payment created", func(t *testing.T) {
		var nonNilPayments []*models.Payment
		for _, p := range results {
			if p != nil {
				nonNilPayments = append(nonNilPayments, p)
			}
		}

		assert.Equal(t, goroutineCount, len(results), "All goroutines should return a result")
		assert.Equal(t, 1, uniquePaymentsCount(nonNilPayments), "Only one unique payment should be created")
	})

	t.Run("Verify no errors returned", func(t *testing.T) {
		for _, err := range errors {
			assert.NoError(t, err)
		}
	})

	t.Run("Verify wallet balance only deducted once", func(t *testing.T) {
		latestWallet, err := tc.WalletRepo.GetByUserId(user.UserID)
		assert.NoError(t, err)

		expectedBalance := wallet.Balance.Sub(req.Amount)
		debugMsg := fmt.Sprintf("Wallet balance should only be deducted once | Wallet Before: %s, Wallet After: %s, Payment Amount: %s",
			wallet.Balance,
			latestWallet.Balance,
			req.Amount.String(),
		)
		assert.True(t, latestWallet.Balance.Equal(expectedBalance), debugMsg)
	})
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
