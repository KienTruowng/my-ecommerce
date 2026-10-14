package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/AiEmStylix/ecommerce/backend/internal/media"
	"github.com/AiEmStylix/ecommerce/backend/internal/user/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type UserHandler struct {
	UserService *UserService
	MediaService *media.MediaService
}

func NewUserHandler(userService *UserService, mediaService *media.MediaService) *UserHandler {
	return &UserHandler{
		UserService: userService,
		MediaService: mediaService,
	}
}

// ==========================================
// 1. INTERNAL / AUTH (No need Token)
// ==========================================

func (u *UserHandler) CreateNewUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Handle logic
	user, err := u.UserService.CreateNewUser(req.Name, req.Email, req.Password)
	
	if err != nil {
		// Return error from service
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Tạo người dùng thành công",
		"data": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Check if email is exists
func (u *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email == "" {
		http.Error(w, "Thiếu email", http.StatusBadRequest)
		return
	}

	user, err := u.UserService.GetUserByEmail(email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// ==========================================
// 2. END-USER / PROFILE (Need Token, ID from Context)
// ==========================================

// Route: GET /api/users/profile
// GetUserProfile godoc
// @Summary      Lấy thông tin Profile
// @Description  Lấy thông tin cá nhân của người dùng đang đăng nhập (dựa vào JWT Token gửi trên Header).
// @Tags         Users
// @ID           user-get-profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserProfileResponse "Thành công, trả về thông tin người dùng"
// @Failure      401  {object}  string                  "Chưa đăng nhập (Token không hợp lệ hoặc bị thiếu)"
// @Failure      404  {object}  string                  "Không tìm thấy người dùng trong hệ thống"
// @Failure      500  {object}  string                  "Lỗi máy chủ nội bộ"
// @Router       /api/users/profile [get]
func (u *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := strconv.Atoi(userIDStr)
	profile, err := u.UserService.GetUserProfileById(uint(userID))

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		return
	}

	// Return data
	utils.WriteJSON(w, http.StatusOK, profile)
}

// Route: PATCH /api/users/profile
// UpdateUserProfile godoc
// @Summary      Cập nhật thông tin User Profile
// @Description  Cập nhật các thông tin cá nhân của người dùng (Name, Email). Bỏ qua các trường bị trống.
// @Tags         Users
// @ID           user-update-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body      body      dto.UpdateProfileRequest      true  "Payload cập nhật thông tin"
// @Success      200       {object}  dto.UserProfileResponse       "Cập nhật thành công, trả về profile mới"
// @Failure      400       {object}  string                        "Dữ liệu gửi lên không hợp lệ"
// @Failure      401       {object}  string                        "Chưa đăng nhập (Token không hợp lệ)"
// @Failure      404       {object}  string                        "Không tìm thấy người dùng"
// @Failure      500       {object}  string                        "Lỗi máy chủ nội bộ"
// @Router       /api/users/profile [patch]
func (u *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// Parse request to dto
	var reqDTO dto.UpdateProfileRequest

	decoder := json.NewDecoder(r.Body)

	// RESTRICT FIELD THAT DOESNT COME FROM DTOs?
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&reqDTO); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	resDTO, err := u.UserService.UpdateUserProfileById(uint(userID), &reqDTO)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "Cannot find user", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resDTO)
}

// Route: POST /api/users/avatar/upload-url
// GetAvatarUploadURL godoc
// @Summary      Lấy link upload avatar
// @Description  Lấy thông tin file để upload avatar lên bucket s3
// @Tags         Users
// @ID           user-get-avatar-upload-url
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body      body      dto.GetPresignedURLRequest    true  "Thông tin file cần upload (tên và định dạng)"
// @Success      200       {object}  dto.GetPresignedURLResponse   "Thành công, trả về link upload và object key"
// @Failure      400       {object}  string                         "Dữ liệu JSON gửi lên không hợp lệ"
// @Failure      401       {object}  string                         "Chưa đăng nhập (Token không hợp lệ)"
// @Failure      500       {object}  string                         "Lỗi máy chủ khi tạo URL"
// @Router       /api/users/avatar/upload-url [post]
func (u *UserHandler) GetAvatarUploadURL(w http.ResponseWriter, r *http.Request) {
	var req dto.GetPresignedURLRequest

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	uploadURL, objectKey, err := u.MediaService.GenerateUploadURL("avatars", req.Filename, req.FileType)

	if err != nil {
        http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
        return
    }

	res := dto.GetPresignedURLResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}

	utils.WriteJSON(w, http.StatusOK, res)
}

// Route: POST /api/users/avatar/confirm
// ConfirmAvatarUpload godoc
// @Summary      Xác nhận upload avatar thành công
// @Description  Sau khi client upload file thành công lên S3/R2 bằng Presigned URL, gọi API này truyền vào object_key để Backend kiểm tra file và lưu URL vào database của User.
// @Tags         Users
// @ID           user-confirm-avatar-upload
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body      body      dto.ConfirmAvatarUploadRequest    true  "Object Key nhận được từ API tạo Upload URL"
// @Success      200       {object}  dto.ConfirmAvatarUploadResponse   "Xác nhận thành công, trả về link avatar mới"
// @Failure      400       {object}  string                            "Dữ liệu không hợp lệ hoặc file vượt quá dung lượng"
// @Failure      401       {object}  string                            "Chưa đăng nhập hoặc token hết hạn"
// @Failure      500       {object}  string                            "Lỗi máy chủ khi cập nhật database"
// @Router       /api/users/avatar/confirm [post]
func (u *UserHandler) ConfirmAvatarUpload(w http.ResponseWriter, r *http.Request) {
	// Sử dụng DTO thay vì inline struct
	var req dto.ConfirmAvatarUploadRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parsedID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid User ID format")
		return
	}

	userID := uint(parsedID)

	// Verify file (giới hạn 5MB)
	finalURL, err := u.MediaService.VerifyUploadFile(req.ObjectKey, 5*1024*1024)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update DB
	err = u.UserService.UpdateUserAvatar(userID, finalURL)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update user avatar in database")
		return
	}

	res := dto.ConfirmAvatarUploadResponse{
		Message:   "Avatar updated successfully",
		AvatarURL: finalURL,
	}

	utils.WriteJSON(w, http.StatusOK, res)
}

// ==========================================
// 3. ADMIN (Admin, require id in url)
// ==========================================

// Route: GET /api/admin/users
// AdminGetUsers godoc
// @Summary      [Admin] Lấy danh sách Users (Có phân trang)
// @Description  Lấy danh sách người dùng cho Admin Dashboard. Hỗ trợ phân trang, tìm kiếm theo tên/email và lọc theo Role.
// @Tags         Admin Users
// @ID           admin-list-users
// @Produce      json
// @Security     BearerAuth
// @Param        page    query     int     false  "Số trang hiện tại (Mặc định: 1)" default(1)
// @Param        limit   query     int     false  "Số lượng bản ghi trên một trang (Mặc định: 10, Tối đa: 100)" default(10)
// @Param        search  query     string  false  "Từ khóa tìm kiếm (Tên hoặc Email)"
// @Param        role    query     string  false  "Lọc theo quyền (Ví dụ: user, admin)"
// @Success      200     {object}  dto.AdminUserListResponse "Danh sách người dùng và thông tin phân trang"
// @Failure      500     {object}  string                    "Lỗi máy chủ nội bộ"
// @Router       /api/admin/users [get]
func (h *UserHandler) AdminGetUsers(w http.ResponseWriter, r *http.Request) {
	// Get params for query
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	keyword := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	users, totalRecords, err := h.UserService.GetAdminUsersList(r.Context(), page, limit, keyword, role)
	if err != nil {
		http.Error(w, "Lỗi khi lấy danh sách người dùng", http.StatusInternalServerError)
		return
	}

	if limit <= 0 { limit = 10 }
	if limit > 100 { limit = 100 }
	if page <= 0 { page = 1 }

	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
	
	meta := dto.PaginationMeta{
        CurrentPage:  page,
        PageSize:     limit,
        TotalRecords: int(totalRecords), // Cast int64 sang int
        TotalPages:   totalPages,
    }
	response := dto.ToAdminUserListResponse(users, meta)
	utils.WriteJSON(w, http.StatusOK, response)
}

// Route: GET /api/admin/users/:id
// AdminGetUserByID godoc
// @Summary      [Admin] Xem chi tiết User
// @Description  Lấy toàn bộ thông tin của một người dùng bao gồm cả role, trạng thái verify và university (Chỉ dành cho Admin).
// @Tags         Admin Users
// @ID           admin-get-user
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID của người dùng"
// @Success      200  {object}  dto.AdminUserProfileResponse "Chi tiết người dùng"
// @Failure      400  {object}  string                       "ID không hợp lệ"
// @Failure      404  {object}  string                       "Không tìm thấy người dùng"
// @Failure      500  {object}  string                       "Lỗi máy chủ nội bộ"
// @Router       /api/admin/users/{id} [get]
func (u *UserHandler) AdminGetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	if idStr == "" {
		http.Error(w, "Thiếu ID người dùng", http.StatusBadRequest)
		return
	}

	idUint64, err := strconv.ParseUint(idStr, 10, 32)

	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	userRes, err := u.UserService.AdminGetUserDetails(uint(idUint64))

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "Không tìm thấy người dùng", http.StatusNotFound)
			return
		}
		http.Error(w, "Lỗi máy chủ", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, userRes)
}

// Route: PATCH /api/admin/users/:id
// AdminUpdateUserByID godoc
// @Summary      [Admin] Cập nhật User
// @Description  Admin thay đổi thông tin người dùng (Role, IsVerified, UniversityID...).
// @Tags         Admin Users
// @ID           admin-update-user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int                          true  "ID của người dùng"
// @Param        body body      dto.AdminUpdateUserRequest   true  "Payload dữ liệu cần update"
// @Success      200  {object}  dto.AdminUserProfileResponse "Thông tin sau khi cập nhật"
// @Failure      400  {object}  string                       "Lỗi dữ liệu"
// @Failure      404  {object}  string                       "Không tìm thấy"
// @Failure      409  {object}  string                       "Email bị trùng"
// @Router       /api/admin/users/{id} [patch]
func (u *UserHandler) AdminUpdateUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Thiếu ID người dùng", http.StatusBadRequest)
		return
	}

	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}
	
	var reqDTO dto.AdminUpdateUserRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&reqDTO); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	resDTO, err := u.UserService.AdminUpdateUser(uint(idUint64), &reqDTO)

	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, "Không tìm thấy người dùng", http.StatusNotFound)
		case errors.Is(err, ErrEmailExists):
			http.Error(w, "Email này đã được sử dụng", http.StatusConflict)
		default:
			http.Error(w, "Lỗi hệ thống", http.StatusInternalServerError)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, resDTO)
}

// Route: DELETE /api/admin/users/:id
// AdminDeleteUserByID godoc
// @Summary      [Admin] Xóa người dùng (Soft Delete)
// @Description  Admin thực hiện xóa mềm một tài khoản người dùng khỏi hệ thống. Tài khoản này sẽ không thể đăng nhập nhưng dữ liệu vẫn còn trong database (deleted_at).
// @Tags         Admin Users
// @ID           admin-delete-user
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID của người dùng cần xóa"
// @Success      204  "Xóa thành công (Không trả về dữ liệu)"
// @Failure      400  {object}  string  "Thiếu ID hoặc ID không hợp lệ"
// @Failure      404  {object}  string  "Người dùng không tồn tại"
// @Failure      500  {object}  string  "Lỗi máy chủ nội bộ"
// @Router       /api/admin/users/{id} [delete]
func (u *UserHandler) AdminDeleteUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Thiếu ID", http.StatusBadRequest)
		return
	}

	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	err = u.UserService.AdminDeleteUser(uint(idUint64))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "Người dùng không tồn tại", http.StatusNotFound)
			return
		}
		http.Error(w, "Lỗi server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Route: POST /api/admin/users/:id/restore
// AdminRestoreUserByID godoc
// @Summary      [Admin] Khôi phục người dùng
// @Description  Admin thực hiện khôi phục tài khoản người dùng đã bị xóa mềm trước đó. Xóa trường deleted_at để user có thể hoạt động lại bình thường.
// @Tags         Admin Users
// @ID           admin-restore-user
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID của người dùng cần khôi phục"
// @Success      204  "Khôi phục thành công (Không trả về dữ liệu)"
// @Failure      400  {object}  string  "ID không hợp lệ hoặc người dùng chưa bị xóa"
// @Failure      404  {object}  string  "Người dùng không tồn tại"
// @Failure      500  {object}  string  "Lỗi máy chủ nội bộ"
// @Router       /api/admin/users/{id}/restore [post]
func (u *UserHandler) AdminRestoreUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	
	if idStr == "" {
		http.Error(w, "Thiếu ID", http.StatusBadRequest)
		return
	}

	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	err = u.UserService.AdminRestoreUser(uint(idUint64))
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, "Người dùng không tồn tại", http.StatusNotFound)
		case errors.Is(err, ErrUserNotDeleted):
			http.Error(w, "Người dùng này chưa bị xóa", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
