package cart

import (
	"encoding/json"
	"errors"
	"log"
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

func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrItemNotFound), errors.Is(err, ErrCartEmpty):
		utils.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrItemNotAvailable), errors.Is(err, ErrSelfCart),
		errors.Is(err, ErrTradeInNotOwned), errors.Is(err, ErrTradeInNotAvail):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
	}
}

func (h *Handler) AddToCart(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body struct {
		ItemID        uint  `json:"item_id"`
		TradeInItemID *uint `json:"offer_item_id,omitempty"`
		CashAmount    uint  `json:"cash_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}
	if body.ItemID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Thiếu item_id")
		return
	}

	log.Printf("[cart] AddToCart userID=%d itemID=%d", callerID, body.ItemID)
	cart, err := h.service.AddToCart(r.Context(), callerID, body.ItemID, body.TradeInItemID, body.CashAmount)
	if err != nil {
		mapError(w, err)
		return
	}

	log.Printf("[cart] AddToCart success — cart ID=%d, item name=%s", cart.ID, cart.Item.Name)
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Đã thêm vào giỏ hàng",
		"data":    cart,
	})
}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cart, err := h.service.GetCart(r.Context(), callerID)
	if err != nil && !errors.Is(err, ErrCartEmpty) {
		mapError(w, err)
		return
	}

	log.Printf("[cart] GetCart userID=%d found=%v", callerID, cart != nil)
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data": cart,
	})
}

func (h *Handler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body struct {
		TradeInItemID *uint `json:"offer_item_id,omitempty"`
		CashAmount    uint  `json:"cash_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	cart, err := h.service.UpdateCart(r.Context(), callerID, body.TradeInItemID, body.CashAmount)
	if err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Đã cập nhật giỏ hàng",
		"data":    cart,
	})
}

func (h *Handler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.RemoveFromCart(r.Context(), callerID); err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Đã xóa khỏi giỏ hàng",
	})
}
