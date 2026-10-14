package offer

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/query"
	"gorm.io/gorm"
)

// --- Sentinel errors ---
var (
	ErrOfferNotFound    = errors.New("Không tìm thấy đề nghị trao đổi")
	ErrItemNotFound     = errors.New("Sản phẩm không tồn tại hoặc đã bị xóa")
	ErrItemNotAvailable = errors.New("Sản phẩm không còn sẵn sàng để trao đổi")
	ErrSelfOffer        = errors.New("Bạn không thể gửi đề nghị cho sản phẩm của chính mình")
	ErrForbidden        = errors.New("Bạn không có quyền thực hiện hành động này")
	ErrInvalidStatus    = errors.New("Trạng thái không hợp lệ, chỉ chấp nhận 'accepted' hoặc 'rejected'")
)

// --- Repository interface ---

type Repository interface {
	Create(ctx context.Context, offer *models.Offer) error
	FindByID(ctx context.Context, id uint) (*models.Offer, error)
	FindSentOffers(ctx context.Context, userID uint) ([]*models.Offer, error)
	FindItemOffers(ctx context.Context, targetItemID uint) ([]*models.Offer, error)
	RejectOffer(ctx context.Context, offerID uint) error
	AcceptOfferTx(ctx context.Context, offer *models.Offer) error
}

// --- Service ---

type Service struct {
	repo Repository
	q    *query.Query
}

func NewService(repo Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, q: query.Use(db)}
}

// CreateOffer: API 1 – Buyer tạo offer mới
func (s *Service) CreateOffer(ctx context.Context, offer *models.Offer) error {
	// Kiểm tra TargetItem tồn tại và available
	targetItem, err := s.q.Item.WithContext(ctx).Where(s.q.Item.ID.Eq(offer.TargetItemID)).First()
	if err != nil {
		return ErrItemNotFound
	}

	if targetItem.Status != models.StatusAvailable {
		return ErrItemNotAvailable
	}

	// Chặn self-offer
	if targetItem.UserID == offer.OfferUserID {
		return ErrSelfOffer
	}

	return s.repo.Create(ctx, offer)
}

// GetSentOffers: API 2 – Buyer xem lịch sử offer đã gửi
func (s *Service) GetSentOffers(ctx context.Context, userID uint) ([]*models.Offer, error) {
	return s.repo.FindSentOffers(ctx, userID)
}

// GetItemOffers: API 3 – Seller xem offer đến item của mình
func (s *Service) GetItemOffers(ctx context.Context, targetItemID uint, callerID uint) ([]*models.Offer, error) {
	// Kiểm tra item tồn tại
	targetItem, err := s.q.Item.WithContext(ctx).Where(s.q.Item.ID.Eq(targetItemID)).First()
	if err != nil {
		return nil, ErrItemNotFound
	}

	// Xác thực quyền sở hữu
	if targetItem.UserID != callerID {
		return nil, ErrForbidden
	}

	return s.repo.FindItemOffers(ctx, targetItemID)
}

// GetOfferForNotify loads an offer with TargetItem and User for notification purposes.
func (s *Service) GetOfferForNotify(ctx context.Context, offerID uint) (*models.Offer, error) {
	return s.repo.FindByID(ctx, offerID)
}

// UpdateOfferStatus: API 4 – Seller duyệt/từ chối offer
func (s *Service) UpdateOfferStatus(ctx context.Context, offerID uint, callerID uint, newStatus models.OfferStatus) error {
	// Validate status đầu vào
	if newStatus != models.OfferStatusAccepted && newStatus != models.OfferStatusRejected {
		return ErrInvalidStatus
	}

	// Tìm offer, preload TargetItem để kiểm tra quyền
	offer, err := s.repo.FindByID(ctx, offerID)
	if err != nil {
		return err
	}

	// Xác thực: chỉ chủ của TargetItem mới được duyệt
	if offer.TargetItem.UserID != callerID {
		return ErrForbidden
	}

	if newStatus == models.OfferStatusRejected {
		return s.repo.RejectOffer(ctx, offerID)
	}

	// accepted – chạy transaction phức tạp
	return s.repo.AcceptOfferTx(ctx, offer)
}

// CancelOffer cancels a pending offer (called by the buyer).
func (s *Service) CancelOffer(ctx context.Context, offerID uint) error {
	_, err := s.q.Offer.WithContext(ctx).
		Where(s.q.Offer.ID.Eq(offerID)).
		Update(s.q.Offer.Status, string(models.OfferStatusCancelled))
	return err
}
