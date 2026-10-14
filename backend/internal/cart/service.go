package cart

import (
	"context"
	"errors"
	"log"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/query"
	"gorm.io/gorm"
)

var (
	ErrItemNotFound     = errors.New("Sản phẩm không tồn tại hoặc đã bị xóa")
	ErrItemNotAvailable = errors.New("Sản phẩm không còn sẵn sàng")
	ErrSelfCart         = errors.New("Bạn không thể thêm sản phẩm của chính mình vào giỏ hàng")
	ErrTradeInNotOwned  = errors.New("Sản phẩm trao đổi phải thuộc về bạn")
	ErrTradeInNotAvail  = errors.New("Sản phẩm trao đổi không còn sẵn sàng")
	ErrCartEmpty        = errors.New("Giỏ hàng trống")
)

type Service struct {
	repo *Repository
	q    *query.Query
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, q: query.Use(db), db: db}
}

func (s *Service) GetCart(ctx context.Context, userID uint) (*models.Cart, error) {
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		log.Printf("[cart.service] GetCart error userID=%d err=%v", userID, err)
		return nil, err
	}
	if cart == nil {
		log.Printf("[cart.service] GetCart empty userID=%d", userID)
		return nil, ErrCartEmpty
	}
	log.Printf("[cart.service] GetCart found userID=%d cartID=%d", userID, cart.ID)
	return cart, nil
}

func (s *Service) AddToCart(ctx context.Context, userID uint, itemID uint, tradeInItemID *uint, cashAmount uint) (*models.Cart, error) {
	var targetItem models.Item
	if err := s.db.WithContext(ctx).
		Preload("ItemImage").
		Where("id = ?", itemID).
		First(&targetItem).Error; err != nil {
		return nil, ErrItemNotFound
	}
	if targetItem.Status != models.StatusAvailable {
		return nil, ErrItemNotAvailable
	}
	if targetItem.UserID == userID {
		return nil, ErrSelfCart
	}

	var tradeInItem *models.Item
	if tradeInItemID != nil {
		var ti models.Item
		if err := s.db.WithContext(ctx).
			Preload("ItemImage").
			Where("id = ?", *tradeInItemID).
			First(&ti).Error; err != nil {
			return nil, ErrItemNotFound
		}
		if ti.UserID != userID {
			return nil, ErrTradeInNotOwned
		}
		if ti.Status != models.StatusAvailable {
			return nil, ErrTradeInNotAvail
		}
		tradeInItem = &ti
	}

	cart := &models.Cart{
		UserID:        userID,
		ItemID:        itemID,
		TradeInItemID: tradeInItemID,
		CashAmount:    cashAmount,
	}
	if err := s.repo.UpsertCart(ctx, cart); err != nil {
		return nil, err
	}

	log.Printf("[cart.service] UpsertCart OK — cartID=%d userID=%d itemID=%d", cart.ID, cart.UserID, cart.ItemID)

	cart.Item = &targetItem
	if tradeInItem != nil {
		cart.TradeInItem = tradeInItem
	}

	return cart, nil
}

func (s *Service) UpdateCart(ctx context.Context, userID uint, tradeInItemID *uint, cashAmount uint) (*models.Cart, error) {
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, ErrCartEmpty
	}

	if tradeInItemID != nil {
		if *tradeInItemID == 0 {
			cart.TradeInItemID = nil
			cart.TradeInItem = nil
		} else {
			var ti models.Item
			if err := s.db.WithContext(ctx).
				Preload("ItemImage").
				Where("id = ?", *tradeInItemID).
				First(&ti).Error; err != nil {
				return nil, ErrItemNotFound
			}
			if ti.UserID != userID {
				return nil, ErrTradeInNotOwned
			}
			if ti.Status != models.StatusAvailable {
				return nil, ErrTradeInNotAvail
			}
			cart.TradeInItemID = tradeInItemID
			cart.TradeInItem = &ti
		}
	}
	cart.CashAmount = cashAmount
	if err := s.repo.UpsertCart(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *Service) RemoveFromCart(ctx context.Context, userID uint) error {
	return s.repo.ClearCart(ctx, userID)
}
