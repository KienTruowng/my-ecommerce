package dto

type UpdateProfileRequest struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	City      string   `json:"city"`
	District  string   `json:"district"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type GetPresignedURLRequest struct {
	Filename string `json:"filename" validate:"required"`
	FileType string `json:"filetype" validate:"required"`
}

type ConfirmAvatarUploadRequest struct {
	ObjectKey string `json:"object_key" validate:"required"`
}

type AdminUpdateUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`        // Admin có quyền đổi role
	IsVerified   bool   `json:"is_verified"` // Admin có quyền verify
	UniversityID *uint  `json:"university_id"`
}
