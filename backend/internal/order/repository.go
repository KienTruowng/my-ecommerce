package order

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) FindByID(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("TradeInItem").
		Preload("Buyer").
		Preload("Seller").
		Preload("Appointment").
		Where("id = ?", orderID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *Repository) FindByBuyerID(ctx context.Context, userID uint) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("TradeInItem").
		Preload("Seller").
		Preload("Appointment").
		Where("buyer_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *Repository) FindBySellerID(ctx context.Context, userID uint) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("TradeInItem").
		Preload("Buyer").
		Preload("Appointment").
		Where("seller_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *Repository) CheckoutTx(ctx context.Context, userID uint, paymentMethod string) (*models.Order, error) {
	var order *models.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		if err := tx.Preload("Item").Preload("Item.User").Preload("TradeInItem").
			Where("user_id = ?", userID).First(&cart).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCartEmpty
			}
			return err
		}

		var item models.Item
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", cart.ItemID).First(&item).Error; err != nil {
			return ErrItemNotFound
		}

		if item.Status != models.StatusAvailable {
			return ErrItemNotAvailable
		}
		if item.UserID == userID {
			return ErrSelfPurchase
		}

		if cart.TradeInItemID != nil {
			var tradeIn models.Item
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", *cart.TradeInItemID).First(&tradeIn).Error; err != nil {
				return ErrItemNotFound
			}
			if tradeIn.UserID != userID {
				return ErrTradeInNotOwned
			}
			if tradeIn.Status != models.StatusAvailable {
				return ErrTradeInNotAvail
			}
		}

		order = &models.Order{
			BuyerID:       userID,
			SellerID:      item.UserID,
			ItemID:        cart.ItemID,
			TradeInItemID: cart.TradeInItemID,
			CashAmount:    cart.CashAmount,
			PaymentMethod: paymentMethod,
			PaymentStatus: models.PaymentStatusPending,
			OrderStatus:   models.OrderStatusConfirmed,
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		if paymentMethod == "wallet" || paymentMethod == "" {
			order.PaymentMethod = "wallet"
			if cart.CashAmount > 0 {
				var wallet models.WalletAccount
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ?", userID).First(&wallet).Error; err != nil {
					return ErrWalletNotFound
				}
				if wallet.Balance < int64(cart.CashAmount) {
					return ErrInsufficientBalance
				}

				externalID := fmt.Sprintf("order-%d-payment", order.ID)
				if err := tx.Create(&models.WalletTransaction{
					WalletAccountID: wallet.ID,
					UserID:          userID,
					Type:            models.WalletTransactionTypeDebit,
					Direction:       models.WalletTransactionDirectionOut,
					Amount:          int64(cart.CashAmount),
					ReferenceType:   "order_payment",
					ReferenceID:     order.ID,
					ExternalID:      externalID,
					Status:          models.WalletTransactionStatusCompleted,
					Metadata:        fmt.Sprintf(`{"order_id":%d}`, order.ID),
				}).Error; err != nil {
					return err
				}

				if err := tx.Model(&wallet).Update("balance", gorm.Expr("balance - ?", cart.CashAmount)).Error; err != nil {
					return err
				}
			}
			order.PaymentStatus = models.PaymentStatusCompleted
			tx.Model(order).Update("payment_status", order.PaymentStatus)
		}

		if err := tx.Model(&models.Item{}).Where("id = ?", cart.ItemID).
			Update("status", string(models.StatusReserved)).Error; err != nil {
			return err
		}

		if cart.TradeInItemID != nil {
			if err := tx.Model(&models.Item{}).Where("id = ?", *cart.TradeInItemID).
				Update("status", string(models.StatusReserved)).Error; err != nil {
				return err
			}
		}

		appt := models.Appointment{
			OrderID:  &order.ID,
			Title:    "Trade meet-up",
			Status:   models.ApptStatusScheduled,
			Location: "",
			Date:     "",
			Time:     "",
		}
		if err := tx.Create(&appt).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", userID).Delete(&models.Cart{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, order.ID)
}

func (r *Repository) CancelOrderTx(ctx context.Context, userID uint, orderID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return ErrOrderNotFound
		}
		if order.BuyerID != userID {
			return ErrForbidden
		}
		if order.OrderStatus != models.OrderStatusConfirmed {
			return ErrOrderNotCancellable
		}

		if order.PaymentStatus == models.PaymentStatusCompleted && order.CashAmount > 0 {
			var wallet models.WalletAccount
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ?", userID).First(&wallet).Error; err != nil {
				return err
			}

			externalID := fmt.Sprintf("order-%d-refund", order.ID)
			if err := tx.Create(&models.WalletTransaction{
				WalletAccountID: wallet.ID,
				UserID:          userID,
				Type:            models.WalletTransactionTypeRefund,
				Direction:       models.WalletTransactionDirectionIn,
				Amount:          int64(order.CashAmount),
				ReferenceType:   "order_refund",
				ReferenceID:     order.ID,
				ExternalID:      externalID,
				Status:          models.WalletTransactionStatusCompleted,
				Metadata:        fmt.Sprintf(`{"order_id":%d}`, order.ID),
			}).Error; err != nil {
				return err
			}

			if err := tx.Model(&wallet).Update("balance", gorm.Expr("balance + ?", order.CashAmount)).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&order).Updates(map[string]any{
			"order_status":   string(models.OrderStatusCancelled),
			"payment_status": string(models.PaymentStatusRefunded),
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Item{}).Where("id = ?", order.ItemID).
			Update("status", string(models.StatusAvailable)).Error; err != nil {
			return err
		}

		if order.TradeInItemID != nil {
			if err := tx.Model(&models.Item{}).Where("id = ?", *order.TradeInItemID).
				Update("status", string(models.StatusAvailable)).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&models.Appointment{}).Where("order_id = ?", order.ID).
			Update("status", string(models.ApptStatusCancelled)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *Repository) CompleteOrderTx(ctx context.Context, orderID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
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
	})
}
