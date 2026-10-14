package auction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrItemNotFound          = errors.New("Sản phẩm không tồn tại hoặc đã bị xóa")
	ErrItemNotAuction        = errors.New("Sản phẩm này không phải loại đấu giá")
	ErrAuctionNotFound       = errors.New("Phiên đấu giá không tồn tại")
	ErrAuctionAlreadyExists  = errors.New("Sản phẩm này đã có cấu hình đấu giá")
	ErrAuctionAlreadySettled = errors.New("Phiên đấu giá đã kết thúc")
	ErrAuctionNotOpen        = errors.New("Phiên đấu giá không còn mở")
	ErrAuctionExpired        = errors.New("Phiên đấu giá đã hết hạn")
	ErrSelfBid               = errors.New("Bạn không thể tự đặt giá cho sản phẩm của mình")
	ErrBidTooLow             = errors.New("Giá đặt phải > giá khởi điểm và > giá cao nhất hiện tại + bước giá")
	ErrStepPriceTooHigh      = errors.New("Bước giá không được vượt quá 20% giá khởi điểm")
	ErrForbidden             = errors.New("Bạn không có quyền thực hiện hành động này")
	ErrDeadlineMustBeFuture  = errors.New("Thời hạn đấu giá phải là thời điểm trong tương lai")
)

// ─── DTOs ────────────────────────────────────────────────────────────────────

type ConfigureAuctionRequest struct {
	BasePrice     uint      `json:"base_price"`
	StepPrice     uint      `json:"step_price"`
	DepositAmount uint      `json:"deposit_amount"`
	Deadline      time.Time `json:"deadline"`
}

type PlaceBidRequest struct {
	Amount uint `json:"amount"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

type Service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// ConfigureAuction: Seller sets up auction parameters for an item they own.
// Business rules enforced:
//   - Item must exist and belong to sellerID
//   - Item must be of type "auction" (or we set it automatically)
//   - StepPrice ≤ 20% of BasePrice
//   - Deadline must be in the future
//   - Only one AuctionConfig per item
func (s *Service) ConfigureAuction(ctx context.Context, sellerID uint, itemID uint, req ConfigureAuctionRequest) (*models.AuctionConfig, error) {
	// Load item
	var item models.Item
	if err := s.db.WithContext(ctx).First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	// Ownership check
	if item.UserID != sellerID {
		return nil, ErrForbidden
	}

	// Deadline must be future
	if req.Deadline.Before(time.Now()) {
		return nil, ErrDeadlineMustBeFuture
	}

	// Validate step price ≤ 20% of base
	maxStep := req.BasePrice / 5 // 20%
	if req.StepPrice > maxStep {
		return nil, ErrStepPriceTooHigh
	}

	// Check existing auction config
	existing, err := s.repo.GetAuctionByItemID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAuctionAlreadyExists
	}

	// Ensure item type is set to auction
	if item.Type != models.TypeAuction {
		s.db.WithContext(ctx).Model(&item).Update("type", models.TypeAuction)
	}

	cfg := &models.AuctionConfig{
		ItemID:        itemID,
		BasePrice:     req.BasePrice,
		StepPrice:     req.StepPrice,
		DepositAmount: req.DepositAmount,
		Deadline:      req.Deadline,
		Status:        models.AuctionStatusOpen,
	}

	if err := s.repo.CreateAuctionConfig(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// GetAuction: Returns auction config + ordered bids for an item.
func (s *Service) GetAuction(ctx context.Context, itemID uint) (*models.AuctionConfig, error) {
	cfg, err := s.repo.GetAuctionByItemID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrAuctionNotFound
	}
	return cfg, nil
}

// PlaceBid: Buyer places a bid. Validates auction is open, not self-bid,
// and amount > currentHighest + stepPrice.
func (s *Service) PlaceBid(ctx context.Context, bidderID uint, itemID uint, req PlaceBidRequest) (*models.Bid, error) {
	// Load auction
	cfg, err := s.repo.GetAuctionByItemID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrAuctionNotFound
	}

	// Auction must be open
	if cfg.Status != models.AuctionStatusOpen {
		return nil, ErrAuctionNotOpen
	}

	// Deadline check
	if time.Now().After(cfg.Deadline) {
		return nil, ErrAuctionExpired
	}

	// Self-bid check
	var item models.Item
	if err := s.db.WithContext(ctx).First(&item, itemID).Error; err != nil {
		return nil, ErrItemNotFound
	}
	if item.UserID == bidderID {
		return nil, ErrSelfBid
	}

	// Fetch current highest bid to calculate minimum
	highestBid, err := s.repo.GetHighestBid(ctx, cfg.ID)
	if err != nil {
		return nil, err
	}

	// Minimum bid: must be strictly above base price AND above current highest + step
	minRequired := cfg.BasePrice
	if highestBid != nil {
		minRequired = highestBid.Amount + cfg.StepPrice
	}

	if req.Amount < minRequired {
		if highestBid == nil {
			return nil, errors.New("giá đặt tối thiểu phải ≥ giá khởi điểm (" + formatVND(cfg.BasePrice) + ")")
		}
		return nil, errors.New("giá đặt tối thiểu phải ≥ giá cao nhất hiện tại (" + formatVND(highestBid.Amount) + ") + bước giá (" + formatVND(cfg.StepPrice) + ") = " + formatVND(minRequired))
	}

	bid := &models.Bid{
		AuctionConfigID: cfg.ID,
		BidderID:        bidderID,
		Amount:          req.Amount,
		Status:          models.BidStatusActive,
	}

	if err := s.repo.PlaceBid(ctx, bid); err != nil {
		return nil, err
	}
	return bid, nil
}

// GetBids: Returns all bids for an item's auction ordered by rank.
func (s *Service) GetBids(ctx context.Context, itemID uint) ([]*models.Bid, error) {
	cfg, err := s.repo.GetAuctionByItemID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrAuctionNotFound
	}
	return s.repo.GetBidsForAuction(ctx, cfg.ID)
}

// SettleAuction: Admin triggers settlement of an expired auction.
// Marks winner (rank-1), marks losers, updates item to traded.
func (s *Service) SettleAuction(ctx context.Context, callerID uint, auctionID uint, isAdmin bool) error {
	if !isAdmin {
		return ErrForbidden
	}

	cfg, err := s.repo.GetAuctionByID(ctx, auctionID)
	if err != nil {
		return err
	}
	if cfg == nil {
		return ErrAuctionNotFound
	}
	if cfg.Status != models.AuctionStatusOpen {
		return ErrAuctionAlreadySettled
	}

	return s.repo.SettleAuctionTx(ctx, auctionID)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func formatVND(amount uint) string {
	if amount == 0 {
		return "0 ₫"
	}
	s := fmt.Sprintf("%d", amount)
	// Insert dots every 3 digits from right
	n := len(s)
	if n <= 3 {
		return s + " ₫"
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result) + " ₫"
}
