package wallet

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var sepayCodePattern = regexp.MustCompile(`SPY-[A-Z0-9]{8}`)

// extractSepayCode extracts the SPY-XXXXXXXX code from the webhook payload.
// SePay may place the code in either the "code" field or the "content" field.
// We check both and return the first match.
func extractSepayCode(txn *models.SepayTransaction) string {
	if txn.Code != nil && *txn.Code != "" {
		if match := sepayCodePattern.FindString(*txn.Code); match != "" {
			return match
		}
	}
	if txn.Content != "" {
		if match := sepayCodePattern.FindString(txn.Content); match != "" {
			return match
		}
	}
	return ""
}

type Repository struct {
	db *gorm.DB
}

var (
	ErrWalletNotFound      = errors.New("Không tìm thấy ví")
	ErrInsufficientBalance = errors.New("Số dư ví không đủ")
)

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTopupRequest(ctx context.Context, req *models.WalletTopupRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *Repository) FindTopupRequestByCode(ctx context.Context, code string) (*models.WalletTopupRequest, error) {
	var req models.WalletTopupRequest
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *Repository) FindWalletByUserID(ctx context.Context, userID uint) (*models.WalletAccount, error) {
	var wallet models.WalletAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *Repository) FindTransactionsByUserID(ctx context.Context, userID uint, limit int) ([]models.WalletTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	var items []models.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

// InsertSepayTransaction inserts a webhook payload with idempotency.
// Uses ON CONFLICT (id) DO NOTHING — retries/replays are safely ignored.
// Returns true if the row was newly inserted, false if it was a duplicate.
func (r *Repository) InsertSepayTransaction(ctx context.Context, txn *models.SepayTransaction) (bool, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(txn)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// MarkSepayTransactionProcessed marks a webhook as successfully processed.
func (r *Repository) MarkSepayTransactionProcessed(ctx context.Context, txnID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SepayTransaction{}).
		Where("id = ?", txnID).
		Updates(map[string]any{
			"processed":    true,
			"processed_at": &now,
		}).Error
}

// ProcessSepayTopupTx handles the full topup flow within a transaction.
// 1. Locks and validates the topup request
// 2. Finds or creates the user's wallet
// 3. Creates a wallet transaction (with external_id idempotency)
// 4. Credits the wallet balance
// 5. Marks the topup request as matched
// Returns nil if the code doesn't match any pending topup (not an error — just a no-op webhook).
func (r *Repository) ProcessSepayTopupTx(ctx context.Context, txn *models.SepayTransaction) error {
	code := extractSepayCode(txn)
	if code == "" {
		return nil
	}
	if txn.TransferType != "in" {
		return nil
	}
	if txn.TransferAmount <= 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var topupReq models.WalletTopupRequest
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", code).
			First(&topupReq).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if topupReq.Status != models.WalletTopupRequestStatusPending {
			return nil
		}
		if topupReq.Amount != txn.TransferAmount {
			return nil
		}

		var wallet models.WalletAccount
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", topupReq.UserID).
			First(&wallet).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			wallet = models.WalletAccount{
				UserID:   topupReq.UserID,
				Balance:  0,
				Currency: "VND",
				Status:   models.WalletStatusActive,
			}
			if err := tx.Create(&wallet).Error; err != nil {
				return err
			}
		}

		externalID := fmt.Sprintf("sepay-%d", txn.ID)
		var dup models.WalletTransaction
		err = tx.Where("external_id = ?", externalID).First(&dup).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Create(&models.WalletTransaction{
			WalletAccountID: wallet.ID,
			UserID:          topupReq.UserID,
			Type:            models.WalletTransactionTypeTopup,
			Direction:       models.WalletTransactionDirectionIn,
			Amount:          txn.TransferAmount,
			ReferenceType:   "sepay_webhook",
			ReferenceID:     topupReq.ID,
			ExternalID:      externalID,
			Status:          models.WalletTransactionStatusCompleted,
			Metadata:        fmt.Sprintf(`{"sepay_webhook_id":%d,"sepay_code":"%s"}`, txn.ID, code),
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&wallet).Update("balance", gorm.Expr("balance + ?", txn.TransferAmount)).Error; err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&topupReq).Updates(map[string]any{
			"status":             models.WalletTopupRequestStatusMatched,
			"matched_webhook_id": txn.ID,
			"matched_at":         &now,
		}).Error
	})
}

func (r *Repository) CreateWalletFromTopupTx(ctx context.Context, topup *models.WalletTopupRequest, webhookID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingTopup models.WalletTopupRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", topup.ID).First(&existingTopup).Error; err != nil {
			return err
		}

		if existingTopup.Status != models.WalletTopupRequestStatusPending {
			return nil
		}

		if existingTopup.Amount <= 0 {
			return nil
		}

		wallet := &models.WalletAccount{}
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", existingTopup.UserID).First(wallet).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			wallet = &models.WalletAccount{UserID: existingTopup.UserID, Balance: 0, Currency: "VND", Status: models.WalletStatusActive}
			if err := tx.Create(wallet).Error; err != nil {
				return err
			}
		}

		var duplicate models.WalletTransaction
		err = tx.Where("external_id = ?", fmt.Sprintf("sepay-%d", webhookID)).First(&duplicate).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Create(&models.WalletTransaction{
			WalletAccountID: wallet.ID,
			UserID:          existingTopup.UserID,
			Type:            models.WalletTransactionTypeTopup,
			Direction:       models.WalletTransactionDirectionIn,
			Amount:          existingTopup.Amount,
			ReferenceType:   "sepay_webhook",
			ReferenceID:     existingTopup.ID,
			ExternalID:      fmt.Sprintf("sepay-%d", webhookID),
			Status:          models.WalletTransactionStatusCompleted,
			Metadata:        fmt.Sprintf(`{"sepay_webhook_id":%d,"sepay_code":"%s"}`, webhookID, existingTopup.Code),
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(wallet).Update("balance", gorm.Expr("balance + ?", existingTopup.Amount)).Error; err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&existingTopup).Updates(map[string]any{
			"status":             models.WalletTopupRequestStatusMatched,
			"matched_webhook_id": webhookID,
			"matched_at":         &now,
		}).Error
	})
}

func (r *Repository) DebitForOfferTx(ctx context.Context, userID uint, amount int64, offerID uint) error {
	if amount <= 0 {
		return nil
	}

	tx := r.db.WithContext(ctx)

	var wallet models.WalletAccount
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWalletNotFound
		}
		return err
	}

	if wallet.Balance < amount {
		return ErrInsufficientBalance
	}

	externalID := fmt.Sprintf("offer-%d-debit", offerID)
	var duplicate models.WalletTransaction
	err := tx.Where("external_id = ?", externalID).First(&duplicate).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := tx.Create(&models.WalletTransaction{
		WalletAccountID: wallet.ID,
		UserID:          userID,
		Type:            models.WalletTransactionTypeDebit,
		Direction:       models.WalletTransactionDirectionOut,
		Amount:          amount,
		ReferenceType:   "offer_cash",
		ReferenceID:     offerID,
		ExternalID:      externalID,
		Status:          models.WalletTransactionStatusCompleted,
		Metadata:        fmt.Sprintf(`{"offer_id":%d}`, offerID),
	}).Error; err != nil {
		return err
	}

	return tx.Model(&wallet).Update("balance", gorm.Expr("balance - ?", amount)).Error
}
