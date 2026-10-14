package user

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUsersPaginated(ctx context.Context, limit, offset int, keyword, role string) ([]models.User, int64, error) {
	var users[]models.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.User{})

	if keyword != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")	
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return users, 0, nil
	}

	err := query.Preload("University").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// Hàm lưu user xuống DB
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.db.Where("email = ?", email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindById(userID uint, preloads ...string) (*models.User, error) {
	var user models.User

	query := r.db

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	err := query.First(&user, userID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	// Update non-zero field
	return r.db.Model(user).Updates(user).Error
}

func (r *UserRepository) SoftDeleteUser(userID uint) error {
	var user models.User
	result := r.db.First(&user, userID)

	if result.Error != nil {
		return result.Error
	}

	result = r.db.Delete(&user)
	return result.Error

}

func (r *UserRepository) FindByIDUnscoped(id uint) (*models.User, error) {
	var user models.User

	err := r.db.Unscoped().
		First(&user, id).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) RestoreDeleteUser(userID uint) (int64, error) {
	result := r.db.Unscoped().
		Model(&models.User{}).
		Where("id = ? AND deleted_at IS NOT NULL", userID).
		Update("deleted_at", nil)

	return result.RowsAffected, result.Error
}

func (r *UserRepository) UpdateAvatar(userID uint, avatarURL string) error {
	return r.db.Model(&models.User{Model: gorm.Model{ID: userID}}).
        Update("avatar_url", avatarURL).Error
}