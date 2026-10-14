package items

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

func NewItemsRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db, q: query.Use(db)}
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*models.Item, error) {
	item, err := r.q.Item.WithContext(ctx).
		Preload(r.q.Item.Categories).
		Preload(r.q.Item.ItemImage).
		Where(r.q.Item.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return item, err
}

func (r *gormRepository) FindByUserID(ctx context.Context, userID uint, status string) ([]*models.Item, error) {
	i := r.q.Item

	stmt := i.WithContext(ctx).
		Preload(i.Categories).
		Preload(i.ItemImage).
		Where(i.UserID.Eq(userID))

	if status != "" {
		stmt = stmt.Where(i.Status.Eq(status))
	}

	items, err := stmt.Find()
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *gormRepository) FindAll(ctx context.Context, categoryID uint, status string, page, limit int) ([]*models.Item, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	stmt := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("ItemImage").
		Preload("User").
		Model(&models.Item{})

	if status != "" {
		stmt = stmt.Where("status = ?", status)
	} else {
		// Default: only show available items in marketplace
		stmt = stmt.Where("status = ?", "available")
	}

	if categoryID > 0 {
		stmt = stmt.Joins("JOIN item_categories ON item_categories.item_id = items.id AND item_categories.category_id = ?", categoryID)
	}

	var total int64
	if err := stmt.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []*models.Item
	if err := stmt.Order("items.created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *gormRepository) Create(ctx context.Context, item *models.Item) error {
	err := r.q.Item.WithContext(ctx).Create(item)

	if err != nil {
		return err
	}

	return nil
}

func (r *gormRepository) Update(ctx context.Context, item *models.Item) error {
	return r.q.Transaction(func(tx *query.Query) error {
		_, err := tx.Item.WithContext(ctx).
			Where(tx.Item.ID.Eq(item.ID)).
			Updates(item)

		if err != nil {
			return err
		}

		// Handle category relationship
		if item.Categories != nil {
			err := tx.Item.Categories.Model(item).Replace(item.Categories...)

			if err != nil {
				return err
			}
		}

		// Handle image relationship
		if item.ItemImage != nil {
			err := tx.Item.ItemImage.Model(item).Replace(item.ItemImage...)

			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormRepository) Delete(ctx context.Context, item *models.Item) error {
	info, err := r.q.Item.Delete(item)

	if err != nil {
		return err
	}

	if info.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *gormRepository) AddImage(ctx context.Context, image *models.ItemImage) error {
	err := r.q.ItemImage.WithContext(ctx).Create(image)
	if err != nil {
		return err
	}
	return nil
}

func (r *gormRepository) RemoveImage(ctx context.Context, imageID uint) error {
	info, err := r.q.ItemImage.WithContext(ctx).
		Where(r.q.ItemImage.ID.Eq(imageID)).
		Delete()
	if err != nil {
		return err
	}

	if info.RowsAffected == 0 {
		return ErrInvalidImage
	}
	return nil
}
