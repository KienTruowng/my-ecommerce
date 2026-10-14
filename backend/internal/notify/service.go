package notify

import (
	"context"
	"fmt"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, userID uint, notifType, title, message, link string) error {
	n := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Link:    link,
	}
	return s.db.WithContext(ctx).Create(n).Error
}

func (s *Service) OfferReceived(ctx context.Context, offer *models.Offer) error {
	if err := s.db.WithContext(ctx).Preload("TargetItem").Preload("User").First(offer, offer.ID).Error; err != nil {
		return err
	}
	itemName := "an item"
	if offer.TargetItem.ID != 0 {
		itemName = offer.TargetItem.Name
	}
	userName := "Someone"
	if offer.User.ID != 0 {
		userName = offer.User.Name
	}

	return s.Create(ctx, offer.TargetItem.UserID, "offer_received",
		"New offer received",
		fmt.Sprintf("%s wants to trade for your \"%s\" — check your offers.", userName, itemName),
		"/offers",
	)
}

func (s *Service) OfferAccepted(ctx context.Context, offer *models.Offer) error {
	if err := s.db.WithContext(ctx).Preload("TargetItem").First(offer, offer.ID).Error; err != nil {
		return err
	}
	itemName := "an item"
	if offer.TargetItem.ID != 0 {
		itemName = offer.TargetItem.Name
	}
	return s.Create(ctx, offer.OfferUserID, "offer_accepted",
		"Offer accepted!",
		fmt.Sprintf("Your offer for \"%s\" was accepted! Go to Appointments to schedule a meetup.", itemName),
		"/appointments",
	)
}

func (s *Service) OfferRejected(ctx context.Context, offer *models.Offer) error {
	if err := s.db.WithContext(ctx).Preload("TargetItem").First(offer, offer.ID).Error; err != nil {
		return err
	}
	itemName := "an item"
	if offer.TargetItem.ID != 0 {
		itemName = offer.TargetItem.Name
	}
	return s.Create(ctx, offer.OfferUserID, "offer_rejected",
		"Offer declined",
		fmt.Sprintf("Your offer for \"%s\" was declined. Try a different item or adjust your offer.", itemName),
		"/offers",
	)
}

func (s *Service) OrderPlaced(ctx context.Context, order *models.Order) error {
	if err := s.db.WithContext(ctx).Preload("Item").Preload("Buyer").First(order, order.ID).Error; err != nil {
		return err
	}
	itemName := "an item"
	if order.Item.ID != 0 {
		itemName = order.Item.Name
	}
	buyerName := "Someone"
	if order.Buyer.ID != 0 {
		buyerName = order.Buyer.Name
	}
	return s.Create(ctx, order.SellerID, "order_placed",
		"New order received",
		fmt.Sprintf("%s placed an order for your \"%s\". Check your orders.", buyerName, itemName),
		"/orders",
	)
}

func (s *Service) OrderCancelled(ctx context.Context, order *models.Order) error {
	if err := s.db.WithContext(ctx).Preload("Item").Preload("Seller").First(order, order.ID).Error; err != nil {
		return err
	}
	itemName := "an item"
	if order.Item.ID != 0 {
		itemName = order.Item.Name
	}
	return s.Create(ctx, order.SellerID, "order_cancelled",
		"Order cancelled",
		fmt.Sprintf("The order for your \"%s\" has been cancelled.", itemName),
		"/orders",
	)
}

func (s *Service) OrderCompleted(ctx context.Context, order *models.Order) error {
	if err := s.db.WithContext(ctx).Preload("Item").Preload("Seller").First(order, order.ID).Error; err != nil {
		return err
	}
	itemName := "an item"
	if order.Item.ID != 0 {
		itemName = order.Item.Name
	}
	return s.Create(ctx, order.BuyerID, "order_completed",
		"Order completed",
		fmt.Sprintf("Your order for \"%s\" has been completed.", itemName),
		"/orders",
	)
}

func (s *Service) GetForUser(ctx context.Context, userID uint, limit int) ([]models.Notification, int64, error) {
	var list []models.Notification
	var unread int64
	s.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&unread)
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, unread, err
}

func (s *Service) MarkRead(ctx context.Context, notifID uint, userID uint) error {
	return s.db.WithContext(ctx).Model(&models.Notification{}).Where("id = ? AND user_id = ?", notifID, userID).Update("is_read", true).Error
}

func (s *Service) MarkAllRead(ctx context.Context, userID uint) error {
	return s.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error
}

func (s *Service) PendingOfferCounts(ctx context.Context, userID uint) (map[uint]int64, error) {
	type row struct {
		TargetItemID uint
		Count        int64
	}
	var rows []row
	err := s.db.WithContext(ctx).Table("offers").
		Select("target_item_id, count(*) as count").
		Joins("JOIN items ON items.id = offers.target_item_id").
		Where("offers.status = ? AND items.user_id = ? AND offers.deleted_at IS NULL", "pending", userID).
		Group("target_item_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]int64)
	for _, r := range rows {
		result[r.TargetItemID] = r.Count
	}
	return result, nil
}
