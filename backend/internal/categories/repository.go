package categories

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

func (r *gormRepository) Create(ctx context.Context, category *models.Category) error {
	err := r.q.Category.WithContext(ctx).Create(category)

	if err != nil {
		return err
	}

	return nil
}

func (r *gormRepository) FindAll(ctx context.Context, limit, offset int) ([]*models.Category, int64, error) {
	// Get the count of record
	totalRecords, err := r.q.Category.WithContext(ctx).Count()

	if err != nil {
		return nil, 0, err
	}

	if totalRecords == 0 {
		return []*models.Category{}, 0, nil
	}

	categories, err := r.q.Category.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order(r.q.Category.ID.Desc()).
		Find()

	if err != nil {
		return nil, 0, err
	}

	return categories, totalRecords, nil
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*models.Category, error) {
	category, err := r.q.Category.WithContext(ctx).Where(r.q.Category.ID.Eq(id)).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

func (r *gormRepository) Update(ctx context.Context, category *models.Category) error {
	info, err := r.q.Category.WithContext(ctx).Where(r.q.Category.ID.Eq(category.ID)).Updates(category)

	if err != nil {
		return err
	}

	if info.RowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *gormRepository) Delete(ctx context.Context, category *models.Category) error {
	info, err := r.q.Category.WithContext(ctx).Delete(category)

	if err != nil {
		return err
	}

	if info.RowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
