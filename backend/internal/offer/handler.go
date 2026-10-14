package offer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	notify "github.com/AiEmStylix/ecommerce/backend/internal/notify"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type Handler struct {
	service *Service
	notify  *notify.Service
}

func NewHandler(service *Service, notifSvc *notify.Service) *Handler {
	return &Handler{service: service, notify: notifSvc}
}

// parseCallerID extracts and parses the authenticated user's ID from context.
func parseCallerID(r *http.Request) (uint, bool) {
	idStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// mapServiceError translates sentinel service errors to appropriate HTTP status codes.
func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrOfferNotFound), errors.Is(err, ErrItemNotFound):
		utils.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		utils.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrSelfOffer), errors.Is(err, ErrItemNotAvailable), errors.Is(err, ErrInvalidStatus):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
	}
}

// CreateOffer godoc
// @Summary      Tạo đề nghị trao đổi (Buyer)
// @Description  Gửi một Offer đến item cụ thể. Item đích phải ở trạng thái `available` và không thể gửi offer cho item của chính mình.
// @Tags         Offers
// @ID           offer-create
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     int           true  "ID của item muốn lấy (TargetItem)"
// @Param        body  body     object{offer_item_id=int,cash_amount=int,offer_message=string}  true  "Thông tin offer. offer_item_id: ID item mang đi đổi (null nếu mua bằng tiền). cash_amount: số tiền thêm vào (0 nếu chỉ đổi item). offer_message: lời nhắn."
// @Success      201   {object} map[string]any "Gửi đề nghị trao đổi thành công"
// @Failure      400   {object} map[string]string "Dữ liệu không hợp lệ | Item không còn available | Tự gửi offer cho chính mình"
// @Failure      401   {object} map[string]string "Unauthorized"
// @Failure      404   {object} map[string]string "Item không tồn tại"
// @Failure      500   {object} map[string]string "Lỗi máy chủ nội bộ"
// @Router       /api/items/{id}/offers [post]
func (h *Handler) CreateOffer(w http.ResponseWriter, r *http.Request) {
	// Parse TargetItemID từ URL
	targetItemIDStr := r.PathValue("id")
	targetItemIDUint64, err := strconv.ParseUint(targetItemIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}
	targetItemID := uint(targetItemIDUint64)

	// Lấy userID của người gọi
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Decode JSON body trực tiếp vào models.Offer (không dùng DTO)
	var offer models.Offer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	// --- Chống Mass Assignment: ghi đè bắt buộc các trường nhạy cảm ---
	offer.OfferUserID = callerID
	offer.TargetItemID = targetItemID
	offer.Status = models.OfferStatusPending

	if err := h.service.CreateOffer(r.Context(), &offer); err != nil {
		mapServiceError(w, err)
		return
	}

	// Notify the item owner
	if h.notify != nil {
		go h.notify.OfferReceived(r.Context(), &offer)
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Gửi đề nghị trao đổi thành công",
		"data":    offer,
	})
}

// GetSentOffers godoc
// @Summary      Lịch sử offer đã gửi (Buyer)
// @Description  Trả về danh sách tất cả các Offer mà người dùng đang đăng nhập đã gửi đi. Bao gồm thông tin đầy đủ về item muốn lấy (TargetItem) và item mang đi đổi (OfferItem).
// @Tags         Offers
// @Produce      json
// @ID           offer-list-sent
// @Security     BearerAuth
// @Success      200  {object}  map[string]any        "Lấy lịch sử đề nghị thành công"
// @Failure      401  {object}  map[string]string     "Unauthorized"
// @Failure      500  {object}  map[string]string     "Lỗi máy chủ nội bộ"
// @Router       /api/users/me/offers [get]
func (h *Handler) GetSentOffers(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	offers, err := h.service.GetSentOffers(r.Context(), callerID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Lấy lịch sử đề nghị thành công",
		"data":    offers,
	})
}

// GetItemOffers godoc
// @Summary      Xem offer đến item của mình (Seller)
// @Description  Trả về danh sách tất cả các Offer gửi đến một item cụ thể. Chỉ chủ sở hữu của item mới có quyền gọi API này. Bao gồm thông tin người gửi (User) và item mang đi đổi (OfferItem).
// @Tags         Offers
// @ID           offer-list-received
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID của TargetItem (item đang được đề nghị)"
// @Success      200  {object}  map[string]any        "Lấy danh sách đề nghị thành công"
// @Failure      400  {object}  map[string]string     "ID không hợp lệ"
// @Failure      401  {object}  map[string]string     "Unauthorized"
// @Failure      403  {object}  map[string]string     "Không có quyền xem offer của item này"
// @Failure      404  {object}  map[string]string     "Item không tồn tại"
// @Failure      500  {object}  map[string]string     "Lỗi máy chủ nội bộ"
// @Router       /api/items/{id}/offers [get]
func (h *Handler) GetItemOffers(w http.ResponseWriter, r *http.Request) {
	targetItemIDStr := r.PathValue("id")
	targetItemIDUint64, err := strconv.ParseUint(targetItemIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}
	targetItemID := uint(targetItemIDUint64)

	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	offers, err := h.service.GetItemOffers(r.Context(), targetItemID, callerID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Lấy danh sách đề nghị thành công",
		"data":    offers,
	})
}

// UpdateOfferStatus godoc
// @Summary      Duyệt hoặc từ chối offer (Seller)
// @Description  Cập nhật trạng thái của một Offer. Chỉ chủ sở hữu TargetItem mới có quyền thực hiện. Khi `accepted`: offer được duyệt, tất cả offer pending còn lại của item đó tự động bị `rejected`, item chuyển sang `traded`. Khi `rejected`: chỉ offer này bị từ chối.
// @Tags         Offers
// @Accept       json
// @ID           offer-update-status
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     int                        true  "ID của Offer cần cập nhật"
// @Param        body  body     object{status=string}      true  "Trạng thái mới: \"accepted\" hoặc \"rejected\""
// @Success      200   {object} map[string]string          "Cập nhật trạng thái đề nghị thành công"
// @Failure      400   {object} map[string]string          "Dữ liệu không hợp lệ | Status không được phép"
// @Failure      401   {object} map[string]string          "Unauthorized"
// @Failure      403   {object} map[string]string          "Không có quyền duyệt offer này"
// @Failure      404   {object} map[string]string          "Offer không tồn tại"
// @Failure      500   {object} map[string]string          "Lỗi máy chủ nội bộ"
// @Router       /api/offers/{id}/status [patch]
func (h *Handler) UpdateOfferStatus(w http.ResponseWriter, r *http.Request) {
	offerIDStr := r.PathValue("id")
	offerIDUint64, err := strconv.ParseUint(offerIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID đề nghị không hợp lệ")
		return
	}
	offerID := uint(offerIDUint64)

	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Decode payload: chỉ cần trường "status"
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	newStatus := models.OfferStatus(payload.Status)

	if err := h.service.UpdateOfferStatus(r.Context(), offerID, callerID, newStatus); err != nil {
		mapServiceError(w, err)
		return
	}

	// Notify the buyer
	if h.notify != nil {
		if foundOffer, err := h.service.GetOfferForNotify(r.Context(), offerID); err == nil {
			if newStatus == models.OfferStatusAccepted {
				go h.notify.OfferAccepted(r.Context(), foundOffer)
			} else {
				go h.notify.OfferRejected(r.Context(), foundOffer)
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Cập nhật trạng thái đề nghị thành công",
	})
}

// CancelOffer allows the buyer to cancel their own pending offer.
func (h *Handler) CancelOffer(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	offerIDStr := r.PathValue("id")
	offerID, err := strconv.ParseUint(offerIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	offer, err := h.service.GetOfferForNotify(r.Context(), uint(offerID))
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Không tìm thấy đề nghị")
		return
	}

	// Only the offer creator can cancel
	if offer.OfferUserID != callerID {
		utils.WriteError(w, http.StatusForbidden, "Bạn không có quyền huỷ đề nghị này")
		return
	}

	// Only pending offers can be cancelled
	if offer.Status != models.OfferStatusPending {
		utils.WriteError(w, http.StatusBadRequest, "Chỉ có thể huỷ đề nghị đang chờ xử lý")
		return
	}

	// Mark as cancelled (we use rejected status for now)
	if err := h.service.CancelOffer(r.Context(), uint(offerID)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi khi huỷ đề nghị")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Đã huỷ đề nghị",
	})
}
