package services

import (
	"errors"
	"fmt"
	"math/rand"
	"payment-service/internal/models"
	"payment-service/internal/redis"
	"payment-service/internal/repositories"
	"payment-service/internal/utils/logger"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentService interface {
	ProcessPayment(ctx *gin.Context, req *models.PaymentRequest) (*models.Payment, error)
	GetPaymentByTransactionID(txId string) (*models.Payment, error)
	GetAll() ([]*models.Payment, error)
}

type paymentService struct {
	logger      logger.Logger
	lockManager *redis.LockManager
	paymentRepo repositories.PaymentRepository
}

func NewPaymentService(paymentRepo repositories.PaymentRepository) PaymentService {
	return &paymentService{
		logger:      logger.Logger{},
		lockManager: redis.NewLockManager(),
		paymentRepo: paymentRepo,
	}
}

func (s *paymentService) ProcessPayment(ctx *gin.Context, req *models.PaymentRequest) (*models.Payment, error) {
	idempotencyKey := req.TransactionID
	if _, ok := s.lockManager.TryLock(idempotencyKey); !ok {
		s.logger.Info("Payment processing, failed to acquired the lock...")
		return s.getOrNil(req)
	}
	defer s.lockManager.Unlock(idempotencyKey)

	time.Sleep(1 * time.Second)

	if exist, err := s.getOrNil(req); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if exist != nil {
		s.logger.Info("found existing payment...")
		return exist, nil
	}

	// Create payment record
	payment := &models.Payment{
		UserID:        req.UserID,
		Amount:        req.Amount,
		TransactionID: req.TransactionID,
		Status:        models.StatusPending,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, err
	}

	// Simulate payment processing
	go s.simulatePaymentProcessing(payment)

	return payment, nil
}

func (s *paymentService) simulatePaymentProcessing(payment *models.Payment) {
	// Simulate processing time (1-3 seconds)
	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

	// Simulate payment success/failure (90% success rate)
	if rand.Float64() < 0.9 {
		payment.Status = models.StatusCompleted
	} else {
		payment.Status = models.StatusFailed
	}

	// Update payment status
	s.paymentRepo.Update(payment)
}

func (s *paymentService) getByTransactionIdAndUserId(payment *models.PaymentRequest) (*models.Payment, error) {
	existing, err := s.paymentRepo.GetByTransactionID(payment.TransactionID)
	if err != nil {
		return nil, err
	}

	if existing.UserID != payment.UserID {
		return nil, fmt.Errorf("user id not match [%s]", payment.UserID)
	}

	return existing, nil
}

func (s *paymentService) getOrNil(req *models.PaymentRequest) (*models.Payment, error) {
	exist, err := s.getByTransactionIdAndUserId(req)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return exist, nil
}

func (s *paymentService) GetPaymentByTransactionID(txId string) (*models.Payment, error) {
	return s.paymentRepo.GetByTransactionID(txId)
}

func (s *paymentService) GetAll() ([]*models.Payment, error) {
	return s.paymentRepo.GetAll()
}
