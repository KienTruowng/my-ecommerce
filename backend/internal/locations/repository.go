package locations

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, limit, offset int, search string, active *bool) ([]*models.Location, int64, error)
	FindByID(ctx context.Context, id uint) (*models.Location, error)
	Create(ctx context.Context, location *models.Location) error
	UpdateFields(ctx context.Context, id uint, updates map[string]any) error
	Delete(ctx context.Context, id uint) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, location *models.Location) error {
	return r.db.WithContext(ctx).Create(location).Error
}

func (r *gormRepository) FindAll(ctx context.Context, limit, offset int, search string, active *bool) ([]*models.Location, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Location{})

	if search != "" {
		query = query.Where("name ILIKE ? OR address ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if active != nil {
		query = query.Where("is_active = ?", *active)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.Location{}, 0, nil
	}

	var locations []*models.Location
	if err := query.Order("updated_at DESC, id DESC").Limit(limit).Offset(offset).Find(&locations).Error; err != nil {
		return nil, 0, err
	}

	return locations, total, nil
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*models.Location, error) {
	var location models.Location

	err := r.db.WithContext(ctx).First(&location, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &location, nil
}

func (r *gormRepository) UpdateFields(ctx context.Context, id uint, updates map[string]any) error {
	result := r.db.WithContext(ctx).Model(&models.Location{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrLocationNotFound
	}
	return nil
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Location{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrLocationNotFound
	}
	return nil
}
