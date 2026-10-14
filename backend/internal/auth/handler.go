package auth

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/AiEmStylix/ecommerce/backend/internal/auth/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type AuthHandler struct {
	AuthService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Register godoc
// @Summary      Đăng ký tài khoản
// @Description  Tạo tài khoản người dùng mới với Name, Email và Password.
// @Tags         Auth
// @ID           auth-register
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterRequest true "Thông tin đăng ký"
// @Success      201  {object} dto.MessageResponse "Đăng ký tài khoản thành công!"
// @Failure      400  {string} string "Dữ liệu không hợp lệ hoặc email đã tồn tại"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Handle logic
	err := h.AuthService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		// Return error from service
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, dto.MessageResponse{
		Message: "Đăng ký tài khoản thành công!",
	})
}

// Login godoc
// @Summary      Đăng nhập
// @Description  Xác thực người dùng bằng Email và Password, trả về JWT Token và thông tin cơ bản.
// @Tags         Auth
// @ID           auth-login
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Thông tin đăng nhập"
// @Success      200  {object} dto.LoginResponse "Đăng nhập thành công!"
// @Failure      400  {string} string "Dữ liệu không hợp lệ"
// @Failure      401  {string} string "Sai email hoặc mật khẩu"
// @Failure      500  {string} string "Lỗi hệ thống"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err:= decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	token, user, err := h.AuthService.Login(req.Email, req.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := dto.LoginResponse{
		Message: "Đăng nhập thành công",
		Token:   string(token),
		User: dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}
	utils.WriteJSON(w, http.StatusOK, response)
}


// ChangePassword godoc
// @Summary      Đổi mật khẩu
// @Description  Người dùng đang đăng nhập thực hiện thay đổi mật khẩu. Yêu cầu truyền lên Access Token ở Header.
// @Tags         Auth
// @ID           auth-change-password
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.UpdatePasswordRequest true "Thông tin mật khẩu cũ và mới"
// @Success      204  "Đổi mật khẩu thành công (Không trả về dữ liệu)"
// @Failure      400  {object} string "Dữ liệu không hợp lệ hoặc sai định dạng"
// @Failure      401  {object} string "Chưa đăng nhập (Token không hợp lệ hoặc bị thiếu)"
// @Failure      500  {object} string "Lỗi máy chủ nội bộ"
// @Router       /api/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {

	userIDStr, ok := appctx.GetUserID(r.Context())

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	var req dto.UpdatePasswordRequest

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	err = h.AuthService.ChangePassword(uint(userID), &req)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
