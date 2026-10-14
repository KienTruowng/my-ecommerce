package items

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/AiEmStylix/ecommerce/backend/internal/items/dto"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
	"gorm.io/gorm"
)

type ItemHandler struct {
	service *ItemsService
}

func NewItemsHandler(service *ItemsService) *ItemHandler {
	return &ItemHandler{service: service}
}

// GetItemByID godoc
// @Summary      Lấy chi tiết sản phẩm
// @Description  Trả về thông tin đầy đủ của một sản phẩm theo ID, bao gồm danh mục và ảnh.
// @Tags         Items
// @ID           item-get
// @Produce      json
// @Param        id   path     int  true  "ID sản phẩm"
// @Success      200  {object} map[string]any    "Thông tin sản phẩm"
// @Failure      400  {object} map[string]string "ID sản phẩm không hợp lệ"
// @Failure      404  {object} map[string]string "Sản phẩm không tồn tại"
// @Failure      500  {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/items/{id} [get]
func (h *ItemHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}

	ctx := r.Context()

	item, err := h.service.FindById(ctx, uint(id))

	// Error Handling
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	// 5. Trả về thành công 200
	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"message": "Lấy thông tin sản phẩm thành công",
		"data":    item,
	}
	json.NewEncoder(w).Encode(response)
}

// BrowseItems godoc
// @Summary      Duyệt sản phẩm trên marketplace
// @Description  Trả về danh sách sản phẩm available của tất cả sinh viên. Hỗ trợ lọc theo category và phân trang.
// @Tags         Items
// @Produce      json
// @ID           item-browse
// @Param        category_id  query    int     false  "Lọc theo category ID"
// @Param        page         query    int     false  "Số trang (mặc định: 1)"
// @Param        limit        query    int     false  "Số lượng mỗi trang (mặc định: 20)"
// @Success      200  {object} dto.BrowseItemsResponse "Danh sách sản phẩm"
// @Failure      500  {object} map[string]string       "Lỗi máy chủ nội bộ"
// @Router       /api/items [get]
func (h *ItemHandler) BrowseItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	categoryIDRaw, _ := strconv.ParseUint(q.Get("category_id"), 10, 32)
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := h.service.BrowseItems(r.Context(), uint(categoryIDRaw), page, limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	data := make([]dto.ItemResponse, 0, len(items))
	for _, item := range items {
		if r := dto.MapToItemResponse(item); r != nil {
			data = append(data, *r)
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	utils.WriteJSON(w, http.StatusOK, dto.BrowseItemsResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			CurrentPage:  page,
			PageSize:     limit,
			TotalRecords: int(total),
			TotalPages:   totalPages,
		},
	})
}

// GetMyItems godoc
// @Summary      Xem kho đồ của tôi
// @Description  Trả về danh sách sản phẩm của người dùng đang đăng nhập. Có thể lọc theo trạng thái (`available`, `reserved`, `traded`, `hidden`).
// @Tags         Items
// @Produce      json
// @ID           item-list-mine
// @Security     BearerAuth
// @Param        status  query    string  false  "Lọc theo trạng thái: available | reserved | traded | hidden"
// @Success      200     {array}  dto.ItemResponse  "Danh sách sản phẩm"
// @Failure      401     {object} map[string]string "Unauthorized"
// @Failure      500     {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/users/me/items [get]
func (h *ItemHandler) GetMyItems(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())

	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in context")
		return
	}

	userIDUint64, err := strconv.ParseUint(userIDStr, 10, 32)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid User ID format")
		return
	}
	userID := uint(userIDUint64)

	status := r.URL.Query().Get("status")

	items, err := h.service.GetMyItems(r.Context(), userID, status)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]dto.ItemResponse, 0)
	for _, item := range items {
		var images []dto.ItemImageResponse
		for _, img := range item.ItemImage {
			images = append(images, dto.ItemImageResponse{
				ID:  img.ID,
				URL: img.URL, // Map tương ứng
			})
		}

		resp = append(resp, dto.ItemResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Type:        string(item.Type),
			Status:      string(item.Status),
			CreatedAt:   item.CreatedAt,
			Images:      images,
		})
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// CreateItem godoc
// @Summary      Đăng sản phẩm mới
// @Description  Tạo một sản phẩm mới gắn với tài khoản đang đăng nhập. Có thể đính kèm danh mục và ảnh.
// @Tags         Items
// @Accept       json
// @Produce      json
// @ID           item-create
// @Security     BearerAuth
// @Param        body  body     dto.CreateItemRequests  true  "Thông tin sản phẩm (name, description, category_ids, image_urls)"
// @Success      201   {object} map[string]any          "Đăng sản phẩm thành công"
// @Failure      400   {object} map[string]string       "Dữ liệu không hợp lệ | Thiếu tên sản phẩm"
// @Failure      401   {object} map[string]string       "Unauthorized"
// @Failure      500   {object} map[string]string       "Lỗi máy chủ nội bộ"
// @Router       /api/items [post]
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())

	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in context")
		return
	}

	userIDUint64, err := strconv.ParseUint(userIDStr, 10, 32)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid User ID format")
		return
	}
	authorizedUserID := uint(userIDUint64)

	var req dto.CreateItemRequests

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tên sản phẩm không được để trống", http.StatusBadRequest)
		return
	}

	newItem := &models.Item{
		Name:        req.Name,
		Description: req.Description,
		UserID:      authorizedUserID,
	}

	for _, catID := range req.CategoryIDs {
		newItem.Categories = append(newItem.Categories, &models.Category{
			Model: gorm.Model{ID: catID},
		})
	}

	for _, imgURL := range req.ImageURLs {
		newItem.ItemImage = append(newItem.ItemImage, &models.ItemImage{
			URL: imgURL,
		})
	}

	err = h.service.CreateNewItem(r.Context(), newItem)

	if err != nil {
		http.Error(w, "Lỗi server khi lưu sản phẩm: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"message": "Tạo sản phẩm thành công",
		"data": map[string]interface{}{
			"id":   newItem.ID, // ID đã được GORM tự động cập nhật sau khi insert thành công
			"name": newItem.Name,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateItem godoc
// @Summary      Cập nhật sản phẩm
// @Description  Cập nhật thông tin sản phẩm (tên, mô tả, danh mục, ảnh) theo ID.
// @Tags         Items
// @Accept       json
// @ID           item-update
// @Produce      json
// @Param        id    path     int                     true  "ID sản phẩm"
// @Param        body  body     dto.UpdateItemsRequest  true  "Thông tin cập nhật"
// @Success      200   {object} map[string]string       "Cập nhật sản phẩm thành công"
// @Failure      400   {object} map[string]string       "Dữ liệu không hợp lệ | Thiếu tên sản phẩm"
// @Failure      404   {object} map[string]string       "Sản phẩm không tồn tại"
// @Failure      500   {object} map[string]string       "Lỗi hệ thống"
// @Router       /api/items/{id} [patch]
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID sản phẩm không hợp lệ", http.StatusBadRequest)
		return
	}

	var req dto.UpdateItemsRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	itemToUpdate := &models.Item{
		Model:       gorm.Model{ID: uint(itemID)},
		Name:        req.Name,
		Description: req.Description,
	}

	if len(req.CategoryIDs) > 0 {
		for _, catID := range req.CategoryIDs {
			itemToUpdate.Categories = append(itemToUpdate.Categories, &models.Category{
				Model: gorm.Model{ID: catID},
			})
		}
	} else {
		itemToUpdate.Categories = []*models.Category{}
	}

	if len(req.ImageURLs) > 0 {
		for _, imgURL := range req.ImageURLs {
			itemToUpdate.ItemImage = append(itemToUpdate.ItemImage, &models.ItemImage{
				URL: imgURL,
			})
		}
	} else {
		itemToUpdate.ItemImage = []*models.ItemImage{}
	}

	err = h.service.UpdateItem(r.Context(), itemToUpdate)

	if err != nil {
		if errors.Is(err, ErrItemsFields) {
			http.Error(w, "Thiếu thông tin bắt buộc: Tên sản phẩm", http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrItemNotFound) {
			http.Error(w, "Sản phẩm không tồn tại", http.StatusNotFound)
		}

		http.Error(w, "Lỗi hệ thống: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Cập nhật sản phẩm thành công",
	})
}

// DeleteItem godoc
// @Summary      Xóa sản phẩm
// @Description  Xóa một sản phẩm theo ID (soft delete).
// @Tags         Items
// @ID           item-delete
// @Produce      json
// @Param        id   path     int  true  "ID sản phẩm"
// @Success      204  "Xóa thành công (không trả về dữ liệu)"
// @Failure      400  {object} map[string]string "ID sản phẩm không hợp lệ"
// @Failure      500  {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/items/{id} [delete]
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.ParseUint(idStr, 10, 32)

	if err != nil {
		http.Error(w, "ID sản phẩm không hợp lệ", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteItemByID(r.Context(), uint(itemID))
	utils.WriteJSON(w, http.StatusNoContent, nil)
}

// RequestUploadImage godoc
// @Summary      Lấy presigned URL để upload ảnh
// @Description  Sinh ra một URL có thời hạn để client upload ảnh trực tiếp lên R2/S3. Sau khi upload xong, gọi API confirm để lưu vào DB.
// @Tags         Items
// @ID           item-presign-image
// @Accept       json
// @Produce      json
// @Param        body  body     dto.PresignImageRequest  true  "Tên file và loại MIME (filename, file_type)"
// @Success      200   {object} map[string]string        "upload_url và object_key"
// @Failure      400   {object} map[string]string        "Dữ liệu không hợp lệ | Thiếu filename hoặc file_type"
// @Failure      500   {object} map[string]string        "Lỗi máy chủ nội bộ"
// @Router       /api/items/images/presign [post]
func (h *ItemHandler) RequestUploadImage(w http.ResponseWriter, r *http.Request) {
	var req dto.PresignImageRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	if req.Filename == "" || req.FileType == "" {
		utils.WriteError(w, http.StatusBadRequest, "Thiếu filename hoặc file_type")
		return
	}

	ctx := r.Context()

	uploadURL, objectKey, err := h.service.RequestUploadImage(ctx, req.Filename, req.FileType)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"upload_url": uploadURL,
		"object_key": objectKey,
	})
}

// ConfirmAndSaveImage godoc
// @Summary      Xác nhận upload ảnh và lưu DB
// @Description  Xác thực file ảnh đã được upload lên R2 (kiểm tra sự tồn tại và dung lượng), sau đó lưu đường dẫn vào cơ sở dữ liệu.
// @Tags         Items
// @ID           item-confirm-image
// @Accept       json
// @Produce      json
// @Param        id    path     int                      true  "ID sản phẩm"
// @Param        body  body     dto.ConfirmImageRequest  true  "object_key trả về từ bước presign"
// @Success      201   {object} map[string]any           "Thêm ảnh thành công"
// @Failure      400   {object} map[string]string        "Dữ liệu không hợp lệ | Ảnh không hợp lệ hoặc quá dung lượng"
// @Failure      404   {object} map[string]string        "Sản phẩm không tồn tại"
// @Failure      500   {object} map[string]string        "Lỗi máy chủ nội bộ"
// @Router       /api/items/{id}/images [post]
func (h *ItemHandler) ConfirmAndSaveImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}

	var req dto.ConfirmImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	if req.ObjectKey == "" {
		utils.WriteError(w, http.StatusBadRequest, "Thiếu object_key")
		return
	}

	// Gọi xuống Service để xác nhận file trên R2 và lưu DB
	newImage, err := h.service.ConfirmAndSaveImage(r.Context(), uint(itemID), req.ObjectKey)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		// Lỗi invalid image hoặc lỗi DB
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Thêm ảnh thành công",
		"data":    newImage,
	})
}

// DeleteImage godoc
// @Summary      Xóa ảnh sản phẩm
// @Description  Xóa một ảnh khỏi DB theo imageID.
// @Tags         Items
// @ID           item-delete-image
// @Produce      json
// @Param        imageID  path     int  true  "ID của ảnh cần xóa"
// @Success      204  "Đã xóa ảnh (không trả về dữ liệu)"
// @Failure      400  {object} map[string]string "ID ảnh không hợp lệ"
// @Failure      500  {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/item-images/{imageID} [delete]
func (h *ItemHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("imageID")
	imageID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID ảnh không hợp lệ")
		return
	}

	err = h.service.DeleteImageByID(r.Context(), uint(imageID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil) // Trả về 204 No Content
}
