package dto

import "time"

type WalletSummaryResponse struct {
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletTransactionResponse struct {
	ID            uint      `json:"id"`
	Type          string    `json:"type"`
	Direction     string    `json:"direction"`
	Amount        int64     `json:"amount"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   uint      `json:"reference_id"`
	ExternalID    string    `json:"external_id"`
	Status        string    `json:"status"`
	Metadata      string    `json:"metadata"`
	CreatedAt     time.Time `json:"created_at"`
}

type WalletTransactionsResponse struct {
	Items []WalletTransactionResponse `json:"items"`
}

type CreateTopupRequestResponse struct {
	Code      string    `json:"code"`
	Amount    int64     `json:"amount"`
	ExpiresAt time.Time `json:"expires_at"`
	Note      string    `json:"note"`
}

type WalletWebhookResponse struct {
	Success bool `json:"success"`
}
