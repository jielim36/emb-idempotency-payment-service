package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	UserID    string          `json:"user_id" gorm:"not null;uniqueIndex"`
	Balance   decimal.Decimal `json:"balance" gorm:"not null"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (wallet *Wallet) Credit(amount decimal.Decimal) {
	wallet.Balance = wallet.Balance.Sub(amount)
}
