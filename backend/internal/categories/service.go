package categories

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, limit, offset int) ([]*models.Category, int64, error)
	FindByID(ctx context.Context, id uint) (*models.Category, error)
	Create(ctx context.Context, category *models.Category) error
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, category *models.Category) error
}

var ErrCategoryNotFound = errors.New("Danh mục không tồn tại hoặc đã bị xóa")
var ErrInvalidCategoryID = errors.New("ID của danh mục không hợp lệ")

type CategoryService struct {
	repo Repository
}

func NewService(repo Repository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (c *CategoryService) CreateNewCategory(ctx context.Context, category *models.Category) (*models.Category, error) {
	err := c.repo.Create(ctx, category)

	if err != nil {
		return nil, err
	}
	return category, nil
}

func (c *CategoryService) FindAll(ctx context.Context, page, limit int) ([]*models.Category, int64, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10 // Default to 10 records
	} else if limit > 100 {
		limit = 100 // Hard limit
	}

	offset := (page - 1) * limit

	categories, totalRecords, err := c.repo.FindAll(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	return categories, totalRecords, nil
}

func (c *CategoryService) FindByID(ctx context.Context, id uint) (*models.Category, error) {
	category, err := c.repo.FindByID(ctx, id)

	// Database error
	if err != nil {
		return nil, err
	}

	// If record doesnt exists
	if category == nil {
		return nil, ErrCategoryNotFound
	}

	return category, nil
}

func (c *CategoryService) UpdateByID(ctx context.Context, id uint, updateData *models.Category) (*models.Category, error) {
	if id == 0 {
		return nil, ErrInvalidCategoryID
	}

	updateData.ID = id

	err := c.repo.Update(ctx, updateData)

	if err != nil {
		return nil, err
	}

	return c.FindByID(ctx, id)
}

func (c *CategoryService) DeleteByID(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidCategoryID
	}

	categoryToDelete := &models.Category{
		Model: gorm.Model{ID: id},
	}

	err := c.repo.Delete(ctx, categoryToDelete)

	if err != nil {
		return err
	}

	return nil
}
