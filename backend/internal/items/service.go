package items

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/media"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.Item, error)
	FindByUserID(ctx context.Context, userID uint, status string) ([]*models.Item, error)
	FindAll(ctx context.Context, categoryID uint, status string, page, limit int) ([]*models.Item, int64, error)
	Create(ctx context.Context, item *models.Item) error
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, item *models.Item) error

	// Method for image
	AddImage(ctx context.Context, image *models.ItemImage) error
	RemoveImage(ctx context.Context, imageID uint) error
}

type ItemsService struct {
	repo  Repository
	media *media.MediaService
}

var ErrItemNotFound = errors.New("Sản phẩm không tồn tại hoặc đã bị xóa")
var ErrItemsFields = errors.New("Lỗi các trường khi cập nhật vật phẩm")
var ErrInvalidID = errors.New("ID sản phẩm không hợp lệ")
var ErrInvalidImage = errors.New("file ảnh không hợp lệ hoặc quá dung lượng")

func NewItemsService(r Repository, m *media.MediaService) *ItemsService {
	return &ItemsService{repo: r, media: m}
}

func (s *ItemsService) FindById(ctx context.Context, id uint) (*models.Item, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, ErrItemNotFound
	}

	return item, nil
}

func (s *ItemsService) GetMyItems(ctx context.Context, userID uint, status string) ([]*models.Item, error) {
	// Chỉ đơn giản là gọi Repo và thực hiện business logic (nếu có)
	return s.repo.FindByUserID(ctx, userID, status)
}

func (s *ItemsService) BrowseItems(ctx context.Context, categoryID uint, page, limit int) ([]*models.Item, int64, error) {
	return s.repo.FindAll(ctx, categoryID, "available", page, limit)
}

func (s *ItemsService) CreateNewItem(ctx context.Context, item *models.Item) error {
	err := s.repo.Create(ctx, item)

	if err != nil {
		return err
	}
	return nil
}

func (s *ItemsService) UpdateItem(ctx context.Context, item *models.Item) error {
	if item.Name == "" {
		return ErrItemsFields
	}

	// Check if item is exists
	existingItem, err := s.repo.FindByID(ctx, item.ID)

	if err != nil {
		return err
	}
	if existingItem == nil {
		return ErrItemNotFound
	}

	return s.repo.Update(ctx, item)
}

func (s *ItemsService) DeleteItemByID(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	itemToDelete := &models.Item{
		Model: gorm.Model{ID: id},
	}

	err := s.repo.Delete(ctx, itemToDelete)

	if err != nil {
		return err
	}

	return nil
}

func (s *ItemsService) RequestUploadImage(ctx context.Context, filename, filetype string) (string, string, error) {
	// Bạn có thể đổi "items" thành thư mục tùy ý trên R2 Bucket
	return s.media.GenerateUploadURL("items", filename, filetype)
}

// ConfirmAndSaveImage:Confirm upload image complete
func (s *ItemsService) ConfirmAndSaveImage(ctx context.Context, itemID uint, objectKey string) (*models.ItemImage, error) {
	// Check if item is exists
	item, err := s.FindById(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}

	// Verify in R2
	finalURL, err := s.media.VerifyUploadFile(objectKey, 5*1024*1024)
	if err != nil {
		return nil, ErrInvalidImage // Che giấu lỗi chi tiết từ media service nếu cần thiết
	}

	newImage := &models.ItemImage{
		ItemID: itemID,
		URL:    finalURL,
	}

	err = s.repo.AddImage(ctx, newImage)
	if err != nil {
		return nil, err
	}

	return newImage, nil
}

// DeleteImageByID: Delete image
func (s *ItemsService) DeleteImageByID(ctx context.Context, imageID uint) error {
	if imageID == 0 {
		return ErrInvalidID
	}
	// Ở một hệ thống thực tế, bạn có thể gọi thêm API của S3/R2 để xóa hẳn file vật lý
	// s.media.DeleteFileFromR2(objectKey) trước khi xóa trong database.

	return s.repo.RemoveImage(ctx, imageID)
}
