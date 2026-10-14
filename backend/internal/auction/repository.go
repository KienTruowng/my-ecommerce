package auction

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

// Repository defines the data-access contract for the auction feature.
type Repository interface {
	// Auction config
	CreateAuctionConfig(ctx context.Context, cfg *models.AuctionConfig) error
	GetAuctionByItemID(ctx context.Context, itemID uint) (*models.AuctionConfig, error)
	GetAuctionByID(ctx context.Context, auctionID uint) (*models.AuctionConfig, error)
	UpdateAuctionStatus(ctx context.Context, auctionID uint, status models.AuctionStatus) error

	// Bids
	PlaceBid(ctx context.Context, bid *models.Bid) error
	GetBidsForAuction(ctx context.Context, auctionID uint) ([]*models.Bid, error)
	GetBidByID(ctx context.Context, bidID uint) (*models.Bid, error)
	GetHighestBid(ctx context.Context, auctionID uint) (*models.Bid, error)

	// Settlement (runs inside a transaction)
	SettleAuctionTx(ctx context.Context, auctionID uint) error
}

// auctionRepository is the concrete gorm-backed implementation.
type auctionRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &auctionRepository{db: db}
}

// ─── AuctionConfig ────────────────────────────────────────────────────────────

func (r *auctionRepository) CreateAuctionConfig(ctx context.Context, cfg *models.AuctionConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *auctionRepository) GetAuctionByItemID(ctx context.Context, itemID uint) (*models.AuctionConfig, error) {
	var cfg models.AuctionConfig
	err := r.db.WithContext(ctx).
		Preload("Bids.Bidder").
		Where("item_id = ?", itemID).
		First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *auctionRepository) GetAuctionByID(ctx context.Context, auctionID uint) (*models.AuctionConfig, error) {
	var cfg models.AuctionConfig
	err := r.db.WithContext(ctx).
		Preload("Bids.Bidder").
		Preload("Item").
		First(&cfg, auctionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *auctionRepository) UpdateAuctionStatus(ctx context.Context, auctionID uint, status models.AuctionStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.AuctionConfig{}).
		Where("id = ?", auctionID).
		Update("status", status).Error
}

// ─── Bids ─────────────────────────────────────────────────────────────────────

func (r *auctionRepository) PlaceBid(ctx context.Context, bid *models.Bid) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reset all existing active bids' rank (will be recalculated below)
		if err := tx.Model(&models.Bid{}).
			Where("auction_config_id = ? AND status = ?", bid.AuctionConfigID, models.BidStatusActive).
			Update("rank", gorm.Expr("rank + 1")).Error; err != nil {
			return err
		}

		// New bid gets rank 1 (highest)
		bid.Rank = 1
		bid.DepositPaid = true // placeholder: assume deposit is always "paid"
		return tx.Create(bid).Error
	})
}

func (r *auctionRepository) GetBidsForAuction(ctx context.Context, auctionID uint) ([]*models.Bid, error) {
	var bids []*models.Bid
	err := r.db.WithContext(ctx).
		Preload("Bidder").
		Where("auction_config_id = ?", auctionID).
		Order("rank ASC, created_at DESC").
		Find(&bids).Error
	return bids, err
}

func (r *auctionRepository) GetBidByID(ctx context.Context, bidID uint) (*models.Bid, error) {
	var bid models.Bid
	err := r.db.WithContext(ctx).Preload("Bidder").First(&bid, bidID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bid, nil
}

func (r *auctionRepository) GetHighestBid(ctx context.Context, auctionID uint) (*models.Bid, error) {
	var bid models.Bid
	err := r.db.WithContext(ctx).
		Where("auction_config_id = ? AND status = ? AND rank = 1", auctionID, models.BidStatusActive).
		First(&bid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bid, nil
}

// ─── Settlement ───────────────────────────────────────────────────────────────

// SettleAuctionTx runs in a single DB transaction:
//  1. Mark the rank-1 bid as "won", all others as "lost"
//  2. Update the AuctionConfig status to "settled"
//  3. Update the Item status to "traded"
func (r *auctionRepository) SettleAuctionTx(ctx context.Context, auctionID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Load auction + bids
		var cfg models.AuctionConfig
		if err := tx.Preload("Bids").First(&cfg, auctionID).Error; err != nil {
			return err
		}

		if cfg.Status != models.AuctionStatusOpen {
			return ErrAuctionAlreadySettled
		}

		// 2. Mark winner (rank 1) and losers
		for _, bid := range cfg.Bids {
			var newStatus models.BidStatus
			if bid.Rank == 1 {
				newStatus = models.BidStatusWon
			} else {
				newStatus = models.BidStatusLost
			}
			if err := tx.Model(bid).Update("status", newStatus).Error; err != nil {
				return err
			}
		}

		// 3. Update auction status
		if err := tx.Model(&cfg).Update("status", models.AuctionStatusSettled).Error; err != nil {
			return err
		}

		// 4. Mark item as traded
		if err := tx.Model(&models.Item{}).Where("id = ?", cfg.ItemID).Update("status", models.StatusTraded).Error; err != nil {
			return err
		}

		return nil
	})
}
