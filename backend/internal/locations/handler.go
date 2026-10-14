package locations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/AiEmStylix/ecommerce/backend/internal/locations/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type LocationHandler struct {
	service *LocationService
}

func NewHandler(service *LocationService) *LocationHandler {
	return &LocationHandler{service: service}
}

// CreateLocation godoc
// @Summary      Tạo địa điểm giao dịch
// @Description  Tạo mới một địa điểm/khu vực giao dịch trong khuôn viên trường.
// @Tags         Locations
// @ID           admin-location-create
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body     dto.CreateLocationRequest  true  "Thông tin địa điểm"
// @Success      201   {object} dto.LocationResponse       "Tạo địa điểm thành công"
// @Failure      400   {object} map[string]string          "Dữ liệu JSON không hợp lệ"
// @Failure      403   {object} map[string]string          "Forbidden: Admin access required"
// @Failure      500   {object} map[string]string          "Lỗi máy chủ nội bộ"
// @Router       /api/admin/locations [post]
func (h *LocationHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLocationRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		utils.WriteError(w, http.StatusBadRequest, "Tên địa điểm không được để trống")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	location := &models.Location{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Address:     strings.TrimSpace(req.Address),
		ImageURL:    strings.TrimSpace(req.ImageURL),
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		OpenTime:    strings.TrimSpace(req.OpenTime),
		CloseTime:   strings.TrimSpace(req.CloseTime),
		IsActive:    isActive,
	}

	created, err := h.service.Create(r.Context(), location)
	if err != nil {
		if errors.Is(err, ErrInvalidLocationName) {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, dto.MapToLocationResponse(created))
}

// GetLocations godoc
// @Summary      Lấy danh sách địa điểm giao dịch
// @Description  Trả về danh sách địa điểm có phân trang, hỗ trợ lọc theo trạng thái và tìm kiếm.
// @Tags         Locations
// @Produce      json
// @ID           location-list
// @Param        page    query    int     false  "Số trang (mặc định: 1)"
// @Param        limit   query    int     false  "Số lượng mỗi trang (mặc định: 10)"
// @Param        search  query    string  false  "Từ khóa tìm theo tên hoặc địa chỉ"
// @Param        active  query    bool    false  "Lọc theo trạng thái hoạt động"
// @Success      200  {object} dto.LocationListResponse "Danh sách địa điểm"
// @Failure      400  {object} map[string]string        "Tham số page/limit không hợp lệ"
// @Failure      500  {object} map[string]string        "Lỗi máy chủ nội bộ"
// @Router       /api/locations [get]
func (h *LocationHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 10

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	activeStr := r.URL.Query().Get("active")

	if pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed < 1 {
			utils.WriteError(w, http.StatusBadRequest, "Invalid 'page' parameter")
			return
		}
		page = parsed
	}

	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.WriteError(w, http.StatusBadRequest, "Invalid 'limit' parameter")
			return
		}
		limit = parsed
	}

	var active *bool
	if activeStr != "" {
		parsed, err := strconv.ParseBool(activeStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid 'active' parameter")
			return
		}
		active = &parsed
	}

	locations, total, err := h.service.List(r.Context(), page, limit, search, active)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	responses := make([]dto.LocationResponse, 0, len(locations))
	for _, location := range locations {
		responses = append(responses, dto.MapToLocationResponse(location))
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	utils.WriteJSON(w, http.StatusOK, dto.LocationListResponse{
		Data: responses,
		Meta: dto.PaginationMeta{
			CurrentPage:  page,
			PageSize:     limit,
			TotalRecords: int(total),
			TotalPages:   totalPages,
		},
	})
}

// GetLocationByID godoc
// @Summary      Lấy chi tiết địa điểm
// @Description  Trả về thông tin chi tiết của một địa điểm theo ID.
// @Tags         Locations
// @Produce      json
// @ID           location-get
// @Param        id   path     int  true  "ID địa điểm"
// @Success      200  {object} dto.LocationResponse "Thông tin địa điểm"
// @Failure      400  {object} map[string]string    "ID không hợp lệ"
// @Failure      404  {object} map[string]string    "Địa điểm không tồn tại"
// @Failure      500  {object} map[string]string    "Lỗi máy chủ nội bộ"
// @Router       /api/locations/{id} [get]
func (h *LocationHandler) GetLocationByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	locationID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID địa điểm không hợp lệ")
		return
	}

	location, err := h.service.GetByID(r.Context(), uint(locationID))
	if err != nil {
		if errors.Is(err, ErrLocationNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	utils.WriteJSON(w, http.StatusOK, dto.MapToLocationResponse(location))
}

// UpdateLocationByID godoc
// @Summary      Cập nhật địa điểm
// @Description  Cập nhật thông tin địa điểm theo ID.
// @Tags         Locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @ID           admin-location-update
// @Param        id    path     int                        true  "ID địa điểm"
// @Param        body  body     dto.UpdateLocationRequest  true  "Thông tin cần cập nhật"
// @Success      200   {object} dto.LocationResponse       "Cập nhật địa điểm thành công"
// @Failure      400   {object} map[string]string          "ID không hợp lệ | Dữ liệu JSON không hợp lệ"
// @Failure      403   {object} map[string]string          "Forbidden: Admin access required"
// @Failure      404   {object} map[string]string          "Địa điểm không tồn tại"
// @Failure      500   {object} map[string]string          "Lỗi máy chủ nội bộ"
// @Router       /api/admin/locations/{id} [patch]
func (h *LocationHandler) UpdateLocationByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	locationID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID địa điểm không hợp lệ")
		return
	}

	var req dto.UpdateLocationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	updated, err := h.service.UpdateByID(r.Context(), uint(locationID), &req)
	if err != nil {
		if errors.Is(err, ErrLocationNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidLocationID) || errors.Is(err, ErrInvalidLocationName) {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	utils.WriteJSON(w, http.StatusOK, dto.MapToLocationResponse(updated))
}

// DeleteLocationByID godoc
// @Summary      Xóa địa điểm
// @Description  Xóa một địa điểm theo ID (soft delete).
// @Tags         Locations
// @Produce      json
// @Security     BearerAuth
// @ID           admin-location-delete
// @Param        id   path     int  true  "ID địa điểm"
// @Success      204  "Xóa thành công (không trả về dữ liệu)"
// @Failure      400  {object} map[string]string "ID không hợp lệ"
// @Failure      403  {object} map[string]string "Forbidden: Admin access required"
// @Failure      404  {object} map[string]string "Địa điểm không tồn tại"
// @Failure      500  {object} map[string]string "Lỗi hệ thống không mong muốn"
// @Router       /api/admin/locations/{id} [delete]
func (h *LocationHandler) DeleteLocationByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	locationID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID địa điểm không hợp lệ")
		return
	}

	err = h.service.DeleteByID(r.Context(), uint(locationID))
	if err != nil {
		if errors.Is(err, ErrLocationNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidLocationID) {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi hệ thống không mong muốn")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
