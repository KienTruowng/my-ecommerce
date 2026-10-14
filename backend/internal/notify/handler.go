package notify

import (
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

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	list, unread, err := h.service.GetForUser(r.Context(), uint(userID), 50)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"notifications": list,
		"unread_count":  unread,
	})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := h.service.MarkRead(r.Context(), uint(id), uint(userID)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)
	if err := h.service.MarkAllRead(r.Context(), uint(userID)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PendingOfferCounts returns how many pending offers each of the user's items has
func (h *Handler) PendingOfferCounts(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	counts, err := h.service.PendingOfferCounts(r.Context(), uint(userID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi")
		return
	}
	utils.WriteJSON(w, http.StatusOK, counts)
}
