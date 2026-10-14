package models

import (
	"time"

	"gorm.io/gorm"
)

type WalletTopupRequestStatus string

const (
	WalletTopupRequestStatusPending WalletTopupRequestStatus = "pending"
	WalletTopupRequestStatusMatched WalletTopupRequestStatus = "matched"
	WalletTopupRequestStatusExpired WalletTopupRequestStatus = "expired"
	WalletTopupRequestStatusFailed  WalletTopupRequestStatus = "failed"
)

type WalletTopupRequest struct {
	gorm.Model

	UserID    uint                     `gorm:"not null;index" json:"user_id"`
	User      User                     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Amount    int64                    `gorm:"not null" json:"amount"`
	Code      string                   `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"`
	Status    WalletTopupRequestStatus `gorm:"type:varchar(20);default:'pending';not null" json:"status"`
	ExpiresAt time.Time                `gorm:"index" json:"expires_at"`

	MatchedWebhookID *int64     `gorm:"index" json:"matched_webhook_id,omitempty"`
	MatchedAt        *time.Time `json:"matched_at,omitempty"`
}
