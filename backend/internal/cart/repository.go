package cart

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetCart(ctx context.Context, userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	cart.Item = &models.Item{}
	r.db.WithContext(ctx).
		Preload("ItemImage").
		Where("id = ?", cart.ItemID).
		First(cart.Item)

	if cart.TradeInItemID != nil {
		var ti models.Item
		r.db.WithContext(ctx).
			Preload("ItemImage").
			Where("id = ?", *cart.TradeInItemID).
			First(&ti)
		cart.TradeInItem = &ti
	}

	return &cart, nil
}

func (r *Repository) UpsertCart(ctx context.Context, cart *models.Cart) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"item_id", "trade_in_item_id", "cash_amount", "updated_at", "deleted_at"}),
		}).
		Create(cart).Error
}

func (r *Repository) ClearCart(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ?", userID).
		Delete(&models.Cart{}).Error
}
