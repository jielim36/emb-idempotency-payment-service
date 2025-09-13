package models

import (
	"time"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	UserID        string        `json:"user_id" gorm:"not null;index" binding:"required"`
	Amount        float64       `json:"amount" gorm:"not null" binding:"required,gt=0"` // can use decimal.Decimal
	TransactionID string        `json:"transaction_id" gorm:"unique;not null;index" binding:"required"`
	Status        PaymentStatus `json:"status" gorm:"default:pending"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type PaymentRequest struct {
	UserID        string  `json:"user_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	TransactionID string  `json:"transaction_id" binding:"required"`
}
