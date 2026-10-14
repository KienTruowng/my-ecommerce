package wallet

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	walletdto "github.com/AiEmStylix/ecommerce/backend/internal/wallet/dto"
)

var (
	ErrInvalidTopupAmount = errors.New("Số tiền nạp không hợp lệ")
	ErrInvalidWebhook     = errors.New("Dữ liệu webhook không hợp lệ")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTopupRequest(ctx context.Context, userID uint, amount int64) (*walletdto.CreateTopupRequestResponse, error) {
	if amount <= 0 {
		return nil, ErrInvalidTopupAmount
	}

	code, err := generateTopupCode()
	if err != nil {
		return nil, err
	}

	request := &models.WalletTopupRequest{
		UserID: userID,
		Amount: amount,
		Code:   code,
		Status: models.WalletTopupRequestStatusPending,
	}
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	request.ExpiresAt = time.Now().Add(24 * time.Hour)

	if err := s.repo.CreateTopupRequest(ctx, request); err != nil {
		return nil, err
	}

	return &walletdto.CreateTopupRequestResponse{
		Code:      code,
		Amount:    amount,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Note:      fmt.Sprintf("Chuyển khoản với mã %s để nạp %d VND", code, amount),
	}, nil
}

func (s *Service) GetWalletSummary(ctx context.Context, userID uint) (*walletdto.WalletSummaryResponse, error) {
	walletAccount, err := s.repo.FindWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if walletAccount == nil {
		return &walletdto.WalletSummaryResponse{
			Balance:   0,
			Currency:  "VND",
			Status:    string(models.WalletStatusActive),
			UpdatedAt: time.Now(),
		}, nil
	}

	return &walletdto.WalletSummaryResponse{
		Balance:   walletAccount.Balance,
		Currency:  walletAccount.Currency,
		Status:    string(walletAccount.Status),
		UpdatedAt: walletAccount.UpdatedAt,
	}, nil
}

func (s *Service) GetWalletTransactions(ctx context.Context, userID uint, limit int) (*walletdto.WalletTransactionsResponse, error) {
	items, err := s.repo.FindTransactionsByUserID(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	resp := make([]walletdto.WalletTransactionResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, walletdto.WalletTransactionResponse{
			ID:            item.ID,
			Type:          string(item.Type),
			Direction:     string(item.Direction),
			Amount:        item.Amount,
			ReferenceType: item.ReferenceType,
			ReferenceID:   item.ReferenceID,
			ExternalID:    item.ExternalID,
			Status:        string(item.Status),
			Metadata:      item.Metadata,
			CreatedAt:     item.CreatedAt,
		})
	}

	return &walletdto.WalletTransactionsResponse{Items: resp}, nil
}

// HandleSePayWebhook is the entry point for SePay webhook processing.
// It follows a strict non-blocking pattern:
//   1. Validate and parse the payload
//   2. Insert into sepay_transactions with ON CONFLICT DO NOTHING (idempotency)
//   3. Return {"success": true} immediately
//   4. Spawn a background goroutine for business logic
//
// The caller (HTTP handler) must respond within 30 seconds per SePay's requirements.
func (s *Service) HandleSePayWebhook(ctx context.Context, payload *walletdto.SePayWebhookPayload) error {
	if payload == nil {
		return ErrInvalidWebhook
	}

	transactionDate, err := time.Parse("2006-01-02 15:04:05", payload.TransactionDate)
	if err != nil {
		return fmt.Errorf("invalid transaction date: %w", err)
	}

	txn := &models.SepayTransaction{
		ID:              payload.ID,
		Gateway:         payload.Gateway,
		TransactionDate: transactionDate,
		AccountNumber:   payload.AccountNumber,
		SubAccount:      payload.SubAccount,
		Code:            payload.Code,
		Content:         payload.Content,
		TransferType:    payload.TransferType,
		Description:     payload.Description,
		TransferAmount:  payload.TransferAmount,
		Accumulated:     payload.Accumulated,
		ReferenceCode:   payload.ReferenceCode,
		ReceivedAt:      time.Now(),
	}

	isNew, err := s.repo.InsertSepayTransaction(ctx, txn)
	if err != nil {
		return fmt.Errorf("failed to insert sepay transaction: %w", err)
	}

	if !isNew {
		log.Printf("[sepay] duplicate webhook ignored: id=%d", payload.ID)
		return nil
	}

	log.Printf("[sepay] new webhook recorded: id=%d gateway=%s amount=%d code=%s",
		payload.ID, payload.Gateway, payload.TransferAmount,
		func() string {
			if payload.Code != nil && *payload.Code != "" {
				return *payload.Code
			}
			return payload.Content
		}())

	// Spawn background goroutine for business logic.
	// This runs after the HTTP handler has already returned 200 OK,
	// ensuring we don't exceed SePay's 30-second response deadline.
	go s.processWebhookInBackground(txn)

	return nil
}

// processWebhookInBackground handles the business logic:
// matching payment codes to pending topup requests and crediting wallets.
// This runs asynchronously in a goroutine.
func (s *Service) processWebhookInBackground(txn *models.SepayTransaction) {
	ctx := context.Background()

	if err := s.repo.ProcessSepayTopupTx(ctx, txn); err != nil {
		log.Printf("[sepay] background processing failed: id=%d err=%v", txn.ID, err)
		return
	}

	if err := s.repo.MarkSepayTransactionProcessed(ctx, txn.ID); err != nil {
		log.Printf("[sepay] failed to mark transaction processed: id=%d err=%v", txn.ID, err)
		return
	}

	log.Printf("[sepay] webhook processed successfully: id=%d content=%s", txn.ID, txn.Content)
}

// HandleSePayWebhookLegacy is kept for backwards compatibility.
// Deprecated: prefer HandleSePayWebhook for the async flow.
func (s *Service) HandleSePayWebhookLegacy(ctx context.Context, payload *walletdto.SePayWebhookPayload) error {
	if payload == nil {
		return ErrInvalidWebhook
	}
	if !strings.EqualFold(payload.TransferType, "in") {
		return nil
	}
	if payload.Code == nil || strings.TrimSpace(*payload.Code) == "" {
		return nil
	}
	if payload.TransferAmount <= 0 {
		return nil
	}

	topupRequest, err := s.repo.FindTopupRequestByCode(ctx, strings.TrimSpace(*payload.Code))
	if err != nil {
		return err
	}
	if topupRequest == nil {
		return nil
	}
	if topupRequest.Status != models.WalletTopupRequestStatusPending {
		return nil
	}
	if topupRequest.Amount != payload.TransferAmount {
		return nil
	}

	return s.repo.CreateWalletFromTopupTx(ctx, topupRequest, payload.ID)
}

func generateTopupCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var builder strings.Builder
	builder.WriteString("SPY-")
	for i := 0; i < 8; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[n.Int64()])
	}
	return builder.String(), nil
}
