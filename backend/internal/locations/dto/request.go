package dto

type CreateLocationRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Address     string   `json:"address"`
	ImageURL    string   `json:"image_url"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	OpenTime    string   `json:"open_time"`
	CloseTime   string   `json:"close_time"`
	IsActive    *bool    `json:"is_active"`
}

type UpdateLocationRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Address     *string  `json:"address"`
	ImageURL    *string  `json:"image_url"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	OpenTime    *string  `json:"open_time"`
	CloseTime   *string  `json:"close_time"`
	IsActive    *bool    `json:"is_active"`
}
