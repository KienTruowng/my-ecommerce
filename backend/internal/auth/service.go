package auth

import (
	"errors"
	"fmt"

	"github.com/AiEmStylix/ecommerce/backend/internal/auth/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindById(id uint, preloads ... string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
}

type AuthService struct {
	userRepo   UserRepository
	jwtService JWTService
}

func NewAuthService(userRepo UserRepository, jwtSvc JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtSvc,
	}
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		return "", nil, errors.New("Lỗi hệ thống khi truy vấn database")
	}

	if user == nil {
		return "", nil, errors.New("Email hoặc mật khẩu không chính xác")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("Email hoặc mật khẩu không chính xác")
	}

	token, err := s.jwtService.GenerateAccessToken(fmt.Sprint(user.ID), user.Role)

	if err != nil {
		return "", nil, errors.New("không thể tạo access token")
	}
	return token, user, nil
}

func (s *AuthService) Register(name, email, password string) error {
	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		return err
	}

	// If user exists
	if user != nil {
		return errors.New("Email đã được đăng ký")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return errors.New("Lỗi khi mã hóa mật khẩu")
	}

	newUser := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	// Lưu user vào database
	err = s.userRepo.Create(newUser)

	if err != nil {
		return errors.New("không thể tạo tài khoản lúc này, vui lòng thử lại sau")
	}

	return nil
}

func (s *AuthService) ChangePassword(id uint, passwordUpdateRequest *dto.UpdatePasswordRequest) error {
	user, err := s.userRepo.FindById(id)

	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("Không tìm thấy người dùng")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordUpdateRequest.OldPassword))

	if err != nil {
		return errors.New("Mật khẩu cũ không chính xác")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(passwordUpdateRequest.NewPassword),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return errors.New("Lỗi khi mã hóa mật khẩu mới")
	}

	user.Password = string(hashedPassword)

	err = s.userRepo.Update(user)

	if err != nil {
		return errors.New("Không thể cập nhật mật khẩu, vui lòng thử lại sau")
	}

	return nil
}
