package appointment

import (
	"context"
	"errors"
	"fmt"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.Appointment, error)
	FindByUserID(ctx context.Context, userID uint) ([]*models.Appointment, error)
	Update(ctx context.Context, appointment *models.Appointment) error
	CompleteAppointmentTx(ctx context.Context, appointment *models.Appointment) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*models.Appointment, error) {
	var appt models.Appointment
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Order.Item").
		Preload("Order.TradeInItem").
		Preload("Order.Buyer").
		Preload("Order.Seller").
		Preload("Offer").
		Preload("Offer.TargetItem").
		Preload("Offer.OfferItem").
		Preload("Offer.User").
		Where("id = ?", id).First(&appt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("appointment not found")
		}
		return nil, err
	}
	return &appt, nil
}

func (r *gormRepository) FindByUserID(ctx context.Context, userID uint) ([]*models.Appointment, error) {
	var appts []*models.Appointment
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Order.Item").
		Preload("Order.TradeInItem").
		Preload("Order.Buyer").
		Preload("Order.Seller").
		Preload("Offer").
		Preload("Offer.TargetItem").
		Preload("Offer.OfferItem").
		Preload("Offer.User").
		Where("appointments.id IN (?)",
			r.db.Table("appointments").
				Select("appointments.id").
				Joins("LEFT JOIN orders ON appointments.order_id = orders.id").
				Joins("LEFT JOIN offers ON appointments.offer_id = offers.id").
				Joins("LEFT JOIN items ON offers.target_item_id = items.id").
				Where("orders.buyer_id = ? OR orders.seller_id = ? OR offers.offer_user_id = ? OR items.user_id = ?",
					userID, userID, userID, userID),
		).
		Order("appointments.created_at DESC").
		Find(&appts).Error
	return appts, err
}

func (r *gormRepository) Update(ctx context.Context, appointment *models.Appointment) error {
	return r.db.WithContext(ctx).Save(appointment).Error
}

func (r *gormRepository) CompleteAppointmentTx(ctx context.Context, appointment *models.Appointment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(appointment).Update("status", string(models.ApptStatusCompleted)).Error; err != nil {
			return err
		}

		if appointment.Order != nil {
			order := appointment.Order

			if err := tx.Model(&models.Item{}).Where("id = ?", order.ItemID).
				Updates(map[string]interface{}{
					"status":  string(models.StatusTraded),
					"user_id": order.BuyerID,
				}).Error; err != nil {
				return err
			}

			if order.TradeInItemID != nil {
				if err := tx.Model(&models.Item{}).Where("id = ?", *order.TradeInItemID).
					Updates(map[string]interface{}{
						"status":  string(models.StatusTraded),
						"user_id": order.SellerID,
					}).Error; err != nil {
					return err
				}
			}

			if order.CashAmount > 0 && order.PaymentStatus == models.PaymentStatusCompleted {
				var sellerWallet models.WalletAccount
				err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ?", order.SellerID).First(&sellerWallet).Error
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					sellerWallet = models.WalletAccount{
						UserID:   order.SellerID,
						Balance:  0,
						Currency: "VND",
						Status:   models.WalletStatusActive,
					}
					if err := tx.Create(&sellerWallet).Error; err != nil {
						return err
					}
				}

				externalID := fmt.Sprintf("order-%d-credit", order.ID)
				if err := tx.Create(&models.WalletTransaction{
					WalletAccountID: sellerWallet.ID,
					UserID:          order.SellerID,
					Type:            models.WalletTransactionTypeTopup,
					Direction:       models.WalletTransactionDirectionIn,
					Amount:          int64(order.CashAmount),
					ReferenceType:   "order_credit",
					ReferenceID:     order.ID,
					ExternalID:      externalID,
					Status:          models.WalletTransactionStatusCompleted,
					Metadata:        fmt.Sprintf(`{"order_id":%d}`, order.ID),
				}).Error; err != nil {
					return err
				}

				if err := tx.Model(&sellerWallet).Update("balance", gorm.Expr("balance + ?", order.CashAmount)).Error; err != nil {
					return err
				}
			}

			return tx.Model(&order).Update("order_status", string(models.OrderStatusCompleted)).Error
		}

		if appointment.Offer != nil {
			offer := appointment.Offer

			if err := tx.Model(&models.Item{}).Where("id = ?", offer.TargetItemID).
				Updates(map[string]interface{}{
					"status":  string(models.StatusTraded),
					"user_id": offer.OfferUserID,
				}).Error; err != nil {
				return err
			}

			if offer.OfferItemID != nil {
				if err := tx.Model(&models.Item{}).Where("id = ?", *offer.OfferItemID).
					Updates(map[string]interface{}{
						"status":  string(models.StatusTraded),
						"user_id": offer.TargetItem.UserID,
					}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
