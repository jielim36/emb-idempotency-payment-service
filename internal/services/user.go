package services

import (
	"payment-service/internal/models"
	"payment-service/internal/repositories"
	"payment-service/internal/utils/logger"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"gorm.io/gorm"
)

type UserService interface {
	Generate() (*models.User, error)
	GetAll() ([]*models.User, error)
	GetByUserId(userId string) (*models.User, error)
	GetUserDetail(userId string) (*models.User, error)
}

type userService struct {
	logger     logger.Logger
	db         *gorm.DB
	userRepo   repositories.UserRepository
	walletRepo repositories.WalletRepository
}

func NewUserService(
	db *gorm.DB,
	userRepo repositories.UserRepository,
	walletRepo repositories.WalletRepository,
) UserService {
	return &userService{
		logger:     logger.Logger{},
		db:         db,
		userRepo:   userRepo,
		walletRepo: walletRepo,
	}
}

// generate a user and wallet for testing used
func (s *userService) Generate() (*models.User, error) {
	DEFAULT_BALANCE := decimal.NewFromInt(10000000) // default 1 million

	user := &models.User{
		UserID: uuid.NewString(), // 自动生成唯一 user_id
	}
	wallet := &models.Wallet{
		UserID:  user.UserID,
		Balance: DEFAULT_BALANCE,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(tx, user); err != nil {
			return err
		}
		if err := s.walletRepo.Create(tx, wallet); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	user.Wallet = wallet
	return user, nil
}

func (s *userService) GetAll() ([]*models.User, error) {
	return s.userRepo.GetAll()
}

func (s *userService) GetByUserId(userId string) (*models.User, error) {
	return s.userRepo.GetByUserId(userId)
}

func (s *userService) GetUserDetail(userId string) (*models.User, error) {
	user, err := s.userRepo.GetByUserId(userId, "Wallet")
	if err != nil {
		return nil, err
	}

	return user, nil
}
