package models

import "gorm.io/gorm"

type WalletStatus string

const (
	WalletStatusActive WalletStatus = "active"
	WalletStatusFrozen WalletStatus = "frozen"
	WalletStatusClosed WalletStatus = "closed"
)

type WalletAccount struct {
	gorm.Model

	UserID   uint         `gorm:"uniqueIndex;not null" json:"user_id"`
	User     User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Balance  int64        `gorm:"not null;default:0" json:"balance"`
	Currency string       `gorm:"type:varchar(3);not null;default:'VND'" json:"currency"`
	Status   WalletStatus `gorm:"type:varchar(20);default:'active';not null" json:"status"`
}
