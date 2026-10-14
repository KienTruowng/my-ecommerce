package dto

import (
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
)

type CategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ItemResponse struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Type        string             `json:"type"`
	Status      string             `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	Images      []ItemImageResponse `json:"images"`
	Categories  []CategoryResponse  `json:"categories"`
	OwnerID     uint               `json:"owner_id"`
}

type ItemImageResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

type BrowseItemsResponse struct {
	Data []ItemResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

// Map dữ liệu từ Model sang Response
func MapToItemResponse(item *models.Item) *ItemResponse {
	var images []ItemImageResponse
	for _, img := range item.ItemImage {
		images = append(images, ItemImageResponse{ID: img.ID, URL: img.URL})
	}
	var cats []CategoryResponse
	for _, c := range item.Categories {
		cats = append(cats, CategoryResponse{ID: c.ID, Name: c.Name})
	}
	return &ItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Type:        string(item.Type),
		Status:      string(item.Status),
		CreatedAt:   item.CreatedAt,
		Images:      images,
		Categories:  cats,
		OwnerID:     item.UserID,
	}
}
