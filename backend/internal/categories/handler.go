package categories

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/categories/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type CategoryHandler struct {
	service *CategoryService
}

func NewHandler(service *CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// CreateNewCategory godoc
// @Summary      Tạo danh mục mới
// @Description  Tạo một danh mục sản phẩm mới. Chỉ yêu cầu trường `name`.
// @Tags         Categories
// @ID           category-create
// @Accept       json
// @Produce      json
// @Param        body  body     dto.CreateCategoryRequest  true  "Tên danh mục"
// @Success      201   {object} map[string]any             "Tạo danh mục thành công"
// @Failure      400   {object} map[string]string          "Dữ liệu JSON không hợp lệ"
// @Failure      500   {object} map[string]string          "Lỗi máy chủ nội bộ"
// @Router       /api/categories [post]
func (h *CategoryHandler) CreateNewCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	// Map the dto to models
	newCategory := &models.Category{
		Name: req.Name,
	}

	createdCategory, err := h.service.CreateNewCategory(r.Context(), newCategory)

	if err != nil {
		http.Error(w, "Lỗi server khi lưu sản phẩm: "+err.Error(), http.StatusInternalServerError)
		return
	}

	categoryDTO := dto.CategoryResponse{
		ID:   createdCategory.ID,
		Name: createdCategory.Name,
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Tạo danh mục thành công",
		"data":    categoryDTO,
	})
}

// GetAllCategory godoc
// @Summary      Lấy danh sách danh mục
// @Description  Trả về danh sách tất cả danh mục có phân trang. Mặc định page=1, limit=10.
// @Tags         Categories
// @Produce      json
// @ID           category-list
// @Param        page   query    int  false  "Số trang (mặc định: 1)"
// @Param        limit  query    int  false  "Số lượng mỗi trang (mặc định: 10)"
// @Success      200    {object} map[string]any    "Danh sách danh mục"
// @Failure      400    {object} map[string]string "Tham số page/limit không hợp lệ"
// @Failure      500    {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/categories [get]
func (h *CategoryHandler) GetAllCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Giá trị mặc định
	page := 1
	limit := 10

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var err error // Khai báo err dùng chung

	// Validation
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			http.Error(w, "Invalid 'page' parameter", http.StatusBadRequest)
			return
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			http.Error(w, "Invalid 'limit' parameter", http.StatusBadRequest)
			return
		}
	}

	categories, total, err := h.service.FindAll(ctx, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create slice with predefined length
	response := make([]dto.CategoryResponse, 0, len(categories))

	for _, c := range categories {
		response = append(response, dto.CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data":  response,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetCategoryByID godoc
// @Summary      Lấy chi tiết danh mục
// @Description  Trả về thông tin chi tiết của một danh mục theo ID.
// @Tags         Categories
// @Produce      json
// @ID           category-get
// @Param        id   path     int  true  "ID danh mục"
// @Success      200  {object} map[string]any    "Thông tin danh mục"
// @Failure      400  {object} map[string]string "ID không hợp lệ"
// @Failure      404  {object} map[string]string "Danh mục không tồn tại"
// @Failure      500  {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	catID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID danh mục không hợp lệ")
		return
	}

	category, err := h.service.FindByID(r.Context(), uint(catID))

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	// Mapping response to dto
	response := dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Lấy thông tin danh mục thành công",
		"data":    response,
	})
}

// UpdateCategoryByID godoc
// @Summary      Cập nhật danh mục
// @Description  Cập nhật tên của một danh mục theo ID.
// @Tags         Categories
// @Accept       json
// @ID           category-update
// @Produce      json
// @Param        id    path     int                        true  "ID danh mục"
// @Param        body  body     dto.UpdateCategoryRequest  true  "Tên danh mục mới"
// @Success      200   {object} map[string]any             "Cập nhật danh mục thành công"
// @Failure      400   {object} map[string]string          "ID không hợp lệ | Dữ liệu JSON không hợp lệ"
// @Failure      404   {object} map[string]string          "Danh mục không tồn tại"
// @Failure      500   {object} map[string]string          "Lỗi hệ thống không mong muốn"
// @Router       /api/categories/{id} [patch]
func (h *CategoryHandler) UpdateCategoryByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	catID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID danh mục không hợp lệ")
		return
	}
	var req dto.UpdateCategoryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	itemToUpdate := &models.Category{
		Name: req.Name,
	}

	updatedCategory, err := h.service.UpdateByID(r.Context(), uint(catID), itemToUpdate)

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, ErrInvalidCategoryID) {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi hệ thống không mong muốn")
		return
	}
	response := dto.CategoryResponse{
		ID:   updatedCategory.ID,
		Name: updatedCategory.Name,
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Cập nhật danh mục thành công",
		"data":    response,
	})
}

// DeleteCategoryByID godoc
// @Summary      Xóa danh mục
// @Description  Xóa một danh mục theo ID (soft delete).
// @Tags         Categories
// @Produce      json
// @Param        id   path     int  true  "ID danh mục"
// @ID           category-delete
// @Success      204  "Xóa thành công (không trả về dữ liệu)"
// @Failure      400  {object} map[string]string "ID không hợp lệ"
// @Failure      404  {object} map[string]string "Danh mục không tồn tại"
// @Failure      500  {object} map[string]string "Lỗi hệ thống không mong muốn"
// @Router       /api/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategoryByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	catID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID danh mục không hợp lệ")
		return
	}

	err = h.service.DeleteByID(r.Context(), uint(catID))

	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, ErrInvalidCategoryID) {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "Lỗi hệ thống không mong muốn")
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}
