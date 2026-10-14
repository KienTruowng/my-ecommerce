package models

import "gorm.io/gorm"

type WalletTransactionType string
type WalletTransactionDirection string
type WalletTransactionStatus string

const (
	WalletTransactionTypeTopup  WalletTransactionType = "topup"
	WalletTransactionTypeDebit  WalletTransactionType = "debit"
	WalletTransactionTypeRefund WalletTransactionType = "refund"
	WalletTransactionTypeAdjust WalletTransactionType = "adjustment"

	WalletTransactionDirectionIn  WalletTransactionDirection = "in"
	WalletTransactionDirectionOut WalletTransactionDirection = "out"

	WalletTransactionStatusPending   WalletTransactionStatus = "pending"
	WalletTransactionStatusCompleted WalletTransactionStatus = "completed"
	WalletTransactionStatusFailed    WalletTransactionStatus = "failed"
	WalletTransactionStatusReversed  WalletTransactionStatus = "reversed"
)

type WalletTransaction struct {
	gorm.Model

	WalletAccountID uint                       `gorm:"not null;index" json:"wallet_account_id"`
	WalletAccount   WalletAccount              `gorm:"foreignKey:WalletAccountID" json:"wallet_account,omitempty"`
	UserID          uint                       `gorm:"not null;index" json:"user_id"`
	User            User                       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type            WalletTransactionType      `gorm:"type:varchar(20);not null" json:"type"`
	Direction       WalletTransactionDirection `gorm:"type:varchar(10);not null" json:"direction"`
	Amount          int64                      `gorm:"not null" json:"amount"`
	ReferenceType   string                     `gorm:"type:varchar(50);not null" json:"reference_type"`
	ReferenceID     uint                       `gorm:"not null;index" json:"reference_id"`
	ExternalID      string                     `gorm:"type:varchar(100);uniqueIndex;not null" json:"external_id"`
	Status          WalletTransactionStatus    `gorm:"type:varchar(20);default:'completed';not null" json:"status"`
	Metadata        string                     `gorm:"type:jsonb;default:'{}'" json:"metadata"`
}
