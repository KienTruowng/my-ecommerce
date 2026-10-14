package dto

import (
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
)

type LocationResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	ImageURL    string    `json:"image_url"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	OpenTime    string    `json:"open_time"`
	CloseTime   string    `json:"close_time"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LocationListResponse struct {
	Data []LocationResponse `json:"data"`
	Meta PaginationMeta     `json:"meta"`
}

type PaginationMeta struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

func MapToLocationResponse(location *models.Location) LocationResponse {
	return LocationResponse{
		ID:          location.ID,
		Name:        location.Name,
		Description: location.Description,
		Address:     location.Address,
		ImageURL:    location.ImageURL,
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
		OpenTime:    location.OpenTime,
		CloseTime:   location.CloseTime,
		IsActive:    location.IsActive,
		CreatedAt:   location.CreatedAt,
		UpdatedAt:   location.UpdatedAt,
	}
}
