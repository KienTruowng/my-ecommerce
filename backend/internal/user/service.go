package user

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/user/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Custom errors
var (
	ErrUserNotFound   = errors.New("User not found")
	ErrUserNotDeleted = errors.New("User is not deleted")
	ErrEmailExists    = errors.New("This email has been used")
)

// ==========================================
// 1. INTERNAL / AUTH (For Login, Register)
// ==========================================

func (u *UserService) CreateNewUser(name, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     name,
		Password: string(hashedPassword),
		Email:    email,
	}

	err = u.repo.Create(user)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) GetUserByEmail(email string) (*models.User, error) {
	user, err := u.repo.FindByEmail(email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Email hoặc mật khẩu không chính xác")
		}
	}
	return user, nil
}

// ==========================================
// 2. END-USER
// ==========================================

func (u *UserService) GetUserProfileById(userId uint) (*dto.UserProfileResponse, error) {
	user, err := u.repo.FindById(userId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	profileResponse := &dto.UserProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		City:      user.City,
		District:  user.District,
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
	}
	return profileResponse, nil
}

func (u *UserService) UpdateUserProfileById(userId uint, profileRequest *dto.UpdateProfileRequest) (*dto.UserProfileResponse, error) {
	user, err := u.repo.FindById(userId)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	// Mapping request DTO to user
	if profileRequest.Name != "" {
		user.Name = profileRequest.Name
	}

	if profileRequest.Email != "" {
		user.Email = profileRequest.Email
	}

	if profileRequest.City != "" {
		user.City = profileRequest.City
	}
	if profileRequest.District != "" {
		user.District = profileRequest.District
	}
	if profileRequest.Latitude != nil {
		user.Latitude = profileRequest.Latitude
	}
	if profileRequest.Longitude != nil {
		user.Longitude = profileRequest.Longitude
	}

	// Update the user profile

	err = u.repo.Update(user)

	// MAPPING MODEL -> DTO RESPONSE
	response := &dto.UserProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		City:      user.City,
		District:  user.District,
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
	}

	return response, nil
}

func (u *UserService) UpdateUserAvatar(userID uint, avatarURL string) error {
	existingUser, err := u.repo.FindById(userID)
	if err != nil {
		return err
	}

	if existingUser == nil {
		return ErrUserNotFound
	}

	err = u.repo.UpdateAvatar(userID, avatarURL)
	if err != nil {
		return errors.New("không thể cập nhật ảnh đại diện vào cơ sở dữ liệu")
	}

	return nil
}

// ==========================================
// 3. ADMIN (Chỉ gọi từ Admin Handlers / Routes)
// ==========================================

func (u *UserService) AdminGetUserDetails(userID uint) (*dto.AdminUserProfileResponse, error) {
	user, err := u.repo.FindById(userID, "University")

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return dto.ToAdminUserProfileResponse(user), nil
}

func (u *UserService) AdminUpdateUser(userID uint, req *dto.AdminUpdateUserRequest) (*dto.AdminUserProfileResponse, error) {
	user, err := u.repo.FindById(userID, "University")
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Email != "" && req.Email != user.Email {
		existing, _ := u.repo.FindByEmail(req.Email)
		if existing != nil {
			return nil, ErrEmailExists
		}
		user.Email = req.Email
	}

	// Map data từ DTO sang Model (Admin được đổi nhiều quyền hơn)
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.UniversityID != nil {
		user.UniversityID = req.UniversityID
	}

	if err := u.repo.Update(user); err != nil {
		return nil, err
	}

	// Get Id to make sure
	updatedUser, _ := u.repo.FindById(userID, "University")

	return dto.ToAdminUserProfileResponse(updatedUser), nil
}

func (u *UserService) AdminDeleteUser(userID uint) error {
	existingUser, err := u.repo.FindById(userID)

	if err != nil {
		return err
	}

	if existingUser == nil {
		return ErrUserNotFound
	}

	return u.repo.SoftDeleteUser(userID)
}

func (u *UserService) AdminRestoreUser(userID uint) error {
	rows, err := u.repo.RestoreDeleteUser(userID)
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotDeleted
	}
	return nil
}

func (u *UserService) GetAdminUsersList(ctx context.Context, page, limit int, keyword, role string) ([]models.User, int64, error) {
	// 1. Validate và set giá trị mặc định
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Hard limit để tránh query quá to làm sập DB
	}

	// 2. Tính toán công thức Offset
	offset := (page - 1) * limit

	// 3. Gọi Repo
	return u.repo.GetUsersPaginated(ctx, limit, offset, keyword, role)
}
