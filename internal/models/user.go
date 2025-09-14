package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Wallet *Wallet `gorm:"foreignKey:UserID;references:UserID"`
}
