package dto

import (
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
)

type UserProfileResponse struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	AvatarURL string   `json:"avatar_url"`
	Role      string   `json:"role"`
	City      string   `json:"city"`
	District  string   `json:"district"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type PublicUserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	City      string `json:"city"` // Only city shown to other users
}

type GetPresignedURLResponse struct {
	UploadURL string `json:"upload_url" example:"https://r2.cloud.com/bucket/..."`
	ObjectKey string `json:"object_key" example:"avatars/171300000_my_avatar.png"`
}

type ConfirmAvatarUploadResponse struct {
	Message   string `json:"message"`
	AvatarURL string `json:"avatar_url"`
}

type AdminUserProfileResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`

	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`

	University *UniversityResponse `json:"university"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UniversityResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PaginationMeta struct {
	CurrentPage  int `json:"current_page" example:"1"`
	PageSize     int `json:"page_size" example:"10"`
	TotalRecords int `json:"total_records" example:"50"`
	TotalPages   int `json:"total_pages" example:"5"`
}

type UserResponse struct {
	ID        uint      `json:"id"` // Khai báo tường minh ở đây
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AdminUserListResponse struct {
	Data []AdminUserProfileResponse `json:"data"`
	Meta PaginationMeta             `json:"meta"`
}

// Helper mapper function
func ToAdminUserProfileResponse(user *models.User) *AdminUserProfileResponse {
	if user == nil {
		return nil
	}

	response := &AdminUserProfileResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		AvatarURL:  user.AvatarURL,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	if user.University != nil {
		response.University = &UniversityResponse{
			ID:   user.University.ID,
			Name: user.University.Name,
		}
	}

	return response
}

func ToAdminUserListResponse(users []models.User, meta PaginationMeta) AdminUserListResponse {
	var data []AdminUserProfileResponse

	for _, user := range users {
		if mappedUser := ToAdminUserProfileResponse(&user); mappedUser != nil {
			data = append(data, *mappedUser)
		}
	}

	if data == nil {
		data = []AdminUserProfileResponse{}
	}

	return AdminUserListResponse{
		Data: data,
		Meta: meta,
	}
}
