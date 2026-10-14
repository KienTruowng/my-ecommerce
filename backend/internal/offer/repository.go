package offer

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/query"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db, q: query.Use(db)}
}

func (r *gormRepository) Create(ctx context.Context, offer *models.Offer) error {
	return r.q.Offer.WithContext(ctx).Create(offer)
}
func (r *gormRepository) FindByID(ctx context.Context, id uint) (*models.Offer, error) {
	offer, err := r.q.Offer.WithContext(ctx).Preload(r.q.Offer.TargetItem).Where(r.q.Offer.ID.Eq(id)).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, err
	}
	return offer, nil
}

// Query offer sent by user (buyer check history)
func (r *gormRepository) FindSentOffers(ctx context.Context, userID uint) ([]*models.Offer, error) {
	o := r.q.Offer

	return o.WithContext(ctx).
		Preload(o.TargetItem).
		Preload(o.OfferItem).
		Where(o.OfferUserID.Eq(userID)).
		Order(o.CreatedAt.Desc()).
		Find()
}

// Query offer receive (seller check history)
func (r *gormRepository) FindItemOffers(ctx context.Context, targetItemID uint) ([]*models.Offer, error) {
	o := r.q.Offer

	return o.WithContext(ctx).
		Preload(o.User).
		Preload(o.OfferItem).
		Where(o.TargetItemID.Eq(targetItemID)).
		Order(o.CreatedAt.Desc()).
		Find()
}

func (r *gormRepository) RejectOffer(ctx context.Context, offerID uint) error {
	_, err := r.q.Offer.WithContext(ctx).
		Where(r.q.Offer.ID.Eq(offerID)).
		Update(r.q.Offer.Status, string(models.OfferStatusRejected))
	return err
}

func (r *gormRepository) AcceptOfferTx(ctx context.Context, offer *models.Offer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qTx := query.Use(tx)

		// 1. Cập nhật offer này thành accepted
		if _, err := qTx.Offer.WithContext(ctx).Where(qTx.Offer.ID.Eq(offer.ID)).Update(qTx.Offer.Status, string(models.OfferStatusAccepted)); err != nil {
			return err
		}

		// 2. Từ chối TẤT CẢ các offer khác của TargetItem này
		if _, err := qTx.Offer.WithContext(ctx).
			Where(qTx.Offer.TargetItemID.Eq(offer.TargetItemID), qTx.Offer.ID.Neq(offer.ID), qTx.Offer.Status.Eq(string(models.OfferStatusPending))).
			Update(qTx.Offer.Status, string(models.OfferStatusRejected)); err != nil {
			return err
		}

		// 3. Cập nhật trạng thái TargetItem thành 'reserved' (không đổi chủ)
		if err := tx.Model(&models.Item{}).WithContext(ctx).
			Where("id = ?", offer.TargetItemID).
			Update("status", string(models.StatusReserved)).Error; err != nil {
			return err
		}

		// 4. Cập nhật trạng thái OfferItem thành 'reserved' nếu có
		if offer.OfferItemID != nil {
			if err := tx.Model(&models.Item{}).WithContext(ctx).
				Where("id = ?", *offer.OfferItemID).
				Update("status", string(models.StatusReserved)).Error; err != nil {
				return err
			}
		}

		// 5. Tự động tạo một Appointment cho hai bên
		appt := models.Appointment{
			OfferID:  &offer.ID,
			Title:    "Trade meet-up",
			Status:   models.ApptStatusScheduled,
			Location: "",
			Date:     "",
			Time:     "",
		}
		if err := tx.WithContext(ctx).Create(&appt).Error; err != nil {
			return err
		}

		return nil
	})
}
