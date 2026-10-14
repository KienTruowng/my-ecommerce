package models

import "time"

// SepayTransaction is the raw webhook log from SePay.
// The `id` field is the PRIMARY KEY and serves as the idempotency key
// — retries/replays with the same id are safely ignored via ON CONFLICT DO NOTHING.
//
// All business logic (wallet crediting, order updates, notifications)
// happens asynchronously in a background goroutine after this record is inserted.
type SepayTransaction struct {
	ID              int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Gateway         string    `gorm:"type:varchar(50)" json:"gateway"`
	TransactionDate time.Time `gorm:"type:timestamp" json:"transaction_date"`
	AccountNumber   string    `gorm:"type:varchar(50)" json:"account_number"`
	SubAccount      string    `gorm:"type:varchar(50)" json:"sub_account"`
	Code            *string   `gorm:"type:varchar(100);index" json:"code"`
	Content         string    `gorm:"type:varchar(500)" json:"content"`
	TransferType    string    `gorm:"type:varchar(10)" json:"transfer_type"`
	Description     string    `gorm:"type:varchar(500)" json:"description"`
	TransferAmount  int64     `gorm:"not null" json:"transfer_amount"`
	Accumulated     int64     `json:"accumulated"`
	ReferenceCode   string    `gorm:"type:varchar(100)" json:"reference_code"`

	Processed   bool      `gorm:"default:false;index" json:"processed"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	ReceivedAt  time.Time `gorm:"autoCreateTime" json:"received_at"`
}
