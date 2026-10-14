package dto

type CreateItemRequests struct {
	Name        string `json:"name"`
	UserID      uint
	Description string   `json:"description"`
	CategoryIDs []uint   `json:"category_ids"`
	ImageURLs   []string `json:"image_urls"`
}

type UpdateItemsRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CategoryIDs []uint   `json:"category_ids"`
	ImageURLs   []string `json:"image_urls"`
}

type PresignImageRequest struct {
	Filename string `json:"filename"`
	FileType string `json:"file_type"` // VD: image/jpeg, image/png
}

type ConfirmImageRequest struct {
	ObjectKey string `json:"object_key"`
}
