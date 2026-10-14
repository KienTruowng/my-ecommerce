package locations

import (
	"context"
	"errors"
	"strings"

	"github.com/AiEmStylix/ecommerce/backend/internal/locations/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
)

type LocationService struct {
	repo Repository
}

var ErrLocationNotFound = errors.New("Địa điểm không tồn tại hoặc đã bị xóa")
var ErrInvalidLocationID = errors.New("ID địa điểm không hợp lệ")
var ErrInvalidLocationName = errors.New("Tên địa điểm không được để trống")

func NewService(repo Repository) *LocationService {
	return &LocationService{repo: repo}
}

func (s *LocationService) Create(ctx context.Context, location *models.Location) (*models.Location, error) {
	if strings.TrimSpace(location.Name) == "" {
		return nil, ErrInvalidLocationName
	}

	err := s.repo.Create(ctx, location)
	if err != nil {
		return nil, err
	}

	return location, nil
}

func (s *LocationService) List(ctx context.Context, page, limit int, search string, active *bool) ([]*models.Location, int64, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	return s.repo.FindAll(ctx, limit, offset, search, active)
}

func (s *LocationService) GetByID(ctx context.Context, id uint) (*models.Location, error) {
	location, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, ErrLocationNotFound
	}
	return location, nil
}

func (s *LocationService) UpdateByID(ctx context.Context, id uint, req *dto.UpdateLocationRequest) (*models.Location, error) {
	if id == 0 {
		return nil, ErrInvalidLocationID
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrLocationNotFound
	}

	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrInvalidLocationName
		}
		updates["name"] = name
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Address != nil {
		updates["address"] = strings.TrimSpace(*req.Address)
	}
	if req.ImageURL != nil {
		updates["image_url"] = strings.TrimSpace(*req.ImageURL)
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}
	if req.OpenTime != nil {
		updates["open_time"] = strings.TrimSpace(*req.OpenTime)
	}
	if req.CloseTime != nil {
		updates["close_time"] = strings.TrimSpace(*req.CloseTime)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return existing, nil
	}

	if err := s.repo.UpdateFields(ctx, id, updates); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *LocationService) DeleteByID(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidLocationID
	}

	return s.repo.Delete(ctx, id)
}
