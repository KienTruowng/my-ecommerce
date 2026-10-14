package order

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
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
	case errors.Is(err, ErrOrderNotFound), errors.Is(err, ErrCartEmpty),
		errors.Is(err, ErrItemNotFound), errors.Is(err, ErrWalletNotFound):
		utils.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		utils.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrItemNotAvailable), errors.Is(err, ErrSelfPurchase),
		errors.Is(err, ErrTradeInNotOwned), errors.Is(err, ErrTradeInNotAvail),
		errors.Is(err, ErrOrderNotCancellable), errors.Is(err, ErrInsufficientBalance):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
	}
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
		return
	}

	order, err := h.service.Checkout(r.Context(), callerID, req.PaymentMethod)
	if err != nil {
		mapError(w, err)
		return
	}

	if h.notify != nil {
		go h.notify.OrderPlaced(r.Context(), order)
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Đặt hàng thành công",
		"data":    ToOrderResponse(order),
	})
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orders, err := h.service.GetOrders(r.Context(), callerID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	resp := make([]*OrderResponse, len(orders))
	for i, o := range orders {
		resp[i] = ToOrderResponse(o)
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data": resp,
	})
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orderIDStr := r.PathValue("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID đơn hàng không hợp lệ")
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), callerID, uint(orderID))
	if err != nil {
		mapError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data": ToOrderResponse(order),
	})
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	callerID, ok := parseCallerID(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orderIDStr := r.PathValue("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID đơn hàng không hợp lệ")
		return
	}

	if err := h.service.CancelOrder(r.Context(), callerID, uint(orderID)); err != nil {
		mapError(w, err)
		return
	}

	if h.notify != nil {
		order, err := h.service.GetOrderByID(r.Context(), callerID, uint(orderID))
		if err == nil {
			go h.notify.OrderCancelled(r.Context(), order)
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Đã hủy đơn hàng",
	})
}
