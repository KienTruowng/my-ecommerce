package auction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// parseCallerID extracts the authenticated user ID from context.
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

// mapError translates service sentinels to HTTP status codes.
func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrItemNotFound), errors.Is(err, ErrAuctionNotFound):
		utils.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		utils.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrAuctionAlreadySettled), errors.Is(err, ErrAuctionAlreadyExists),
		errors.Is(err, ErrSelfBid), errors.Is(err, ErrBidTooLow),
		errors.Is(err, ErrStepPriceTooHigh), errors.Is(err, ErrAuctionExpired),
		errors.Is(err, ErrAuctionNotOpen), errors.Is(err, ErrItemNotAuction),
		errors.Is(err, ErrDeadlineMustBeFuture):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
	}
}

// ConfigureAuction godoc
// @Summary      Cấu hình đấu giá cho sản phẩm
// @Description  Người bán thiết lập các thông số đấu giá: giá khởi điểm, bước giá (tối đa 20% giá khởi điểm), tiền cọc và thời hạn. Chỉ chủ sản phẩm mới có quyền.
// @Tags         Auction
// @ID           auction-configure
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     int                       true  "ID sản phẩm"
// @Param        body  body     ConfigureAuctionRequest   true  "Thông số đấu giá"
// @Success      201   {object} map[string]any            "Cấu hình đấu giá thành công"
// @Failure      400   {object} map[string]string         "Dữ liệu không hợp lệ"
// @Failure      401   {object} map[string]string         "Unauthorized"
// @Failure      403   {object} map[string]string         "Không có quyền"
// @Failure      404   {object} map[string]string         "Sản phẩm không tồn tại"
// @Router       /api/items/{id}/auction [post]
func (h *Handler) ConfigureAuction(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}

	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ConfigureAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	cfg, err := h.service.ConfigureAuction(r.Context(), callerID, itemID, req)
	if err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Cấu hình đấu giá thành công",
		"data":    cfg,
	})
}

// GetAuction godoc
// @Summary      Lấy thông tin phiên đấu giá
// @Description  Trả về cấu hình đấu giá và danh sách các lượt đặt giá (sắp xếp theo thứ hạng) của một sản phẩm.
// @Tags         Auction
// @ID           auction-get
// @Produce      json
// @Param        id   path     int  true  "ID sản phẩm"
// @Success      200  {object} map[string]any            "Thông tin đấu giá"
// @Failure      404  {object} map[string]string         "Phiên đấu giá không tồn tại"
// @Router       /api/items/{id}/auction [get]
func (h *Handler) GetAuction(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}

	cfg, err := h.service.GetAuction(r.Context(), itemID)
	if err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Lấy thông tin đấu giá thành công",
		"data":    cfg,
	})
}

// PlaceBid godoc
// @Summary      Đặt giá trong phiên đấu giá
// @Description  Người mua đặt một mức giá. Giá phải ≥ (giá cao nhất hiện tại + bước giá). Hệ thống tự động ghi nhận tiền cọc (placeholder).
// @Tags         Auction
// @ID           auction-bid
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     int               true  "ID sản phẩm"
// @Param        body  body     PlaceBidRequest   true  "Mức giá đặt"
// @Success      201   {object} map[string]any    "Đặt giá thành công"
// @Failure      400   {object} map[string]string "Giá thấp hơn mức tối thiểu | Phiên đã kết thúc | Tự đặt giá"
// @Failure      401   {object} map[string]string "Unauthorized"
// @Failure      404   {object} map[string]string "Phiên đấu giá không tồn tại"
// @Router       /api/items/{id}/bids [post]
func (h *Handler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}

	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	bid, err := h.service.PlaceBid(r.Context(), callerID, itemID, req)
	if err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Đặt giá thành công",
		"data":    bid,
	})
}

// GetBids godoc
// @Summary      Danh sách lượt đặt giá
// @Description  Trả về tất cả các lượt đặt giá của một phiên đấu giá, sắp xếp theo thứ hạng (cao nhất trước).
// @Tags         Auction
// @ID           auction-bids-list
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     int  true  "ID sản phẩm"
// @Success      200  {object} map[string]any            "Danh sách lượt đặt giá"
// @Failure      404  {object} map[string]string         "Phiên đấu giá không tồn tại"
// @Router       /api/items/{id}/bids [get]
func (h *Handler) GetBids(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID sản phẩm không hợp lệ")
		return
	}

	bids, err := h.service.GetBids(r.Context(), itemID)
	if err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Lấy danh sách lượt đặt giá thành công",
		"data":    bids,
	})
}

// SettleAuction godoc
// @Summary      Kết thúc phiên đấu giá (Admin)
// @Description  Admin thủ công kết thúc một phiên đấu giá: xác nhận người thắng (rank 1), trả lại tiền cọc cho người thua, chuyển trạng thái sản phẩm thành 'traded'.
// @Tags         Auction
// @ID           auction-settle
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     int  true  "ID phiên đấu giá (AuctionConfig ID)"
// @Success      200  {object} map[string]string         "Kết thúc phiên đấu giá thành công"
// @Failure      400  {object} map[string]string         "Phiên đấu giá đã được kết thúc trước đó"
// @Failure      403  {object} map[string]string         "Không có quyền"
// @Failure      404  {object} map[string]string         "Phiên đấu giá không tồn tại"
// @Router       /api/admin/auctions/{id}/settle [post]
func (h *Handler) SettleAuction(w http.ResponseWriter, r *http.Request) {
	auctionIDStr := r.PathValue("id")
	auctionIDUint64, err := strconv.ParseUint(auctionIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID phiên đấu giá không hợp lệ")
		return
	}
	auctionID := uint(auctionIDUint64)

	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// isAdmin=true because this route is protected by requireAdmin middleware
	if err := h.service.SettleAuction(r.Context(), callerID, auctionID, true); err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Kết thúc phiên đấu giá thành công",
	})
}

// parseItemID reads the {id} path value as a uint.
func parseItemID(r *http.Request) (uint, error) {
	raw := r.PathValue("id")
	v, err := strconv.ParseUint(raw, 10, 64)
	return uint(v), err
}
