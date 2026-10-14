package wallet

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
	walletdto "github.com/AiEmStylix/ecommerce/backend/internal/wallet/dto"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetWalletSummary(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parsedUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "User ID không hợp lệ")
		return
	}

	resp, err := h.service.GetWalletSummary(r.Context(), uint(parsedUserID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parsedUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "User ID không hợp lệ")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, parseErr := strconv.Atoi(limitStr)
		if parseErr != nil || parsedLimit <= 0 {
			utils.WriteError(w, http.StatusBadRequest, "limit không hợp lệ")
			return
		}
		limit = parsedLimit
	}

	resp, err := h.service.GetWalletTransactions(r.Context(), uint(parsedUserID), limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateTopupRequest(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parsedUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "User ID không hợp lệ")
		return
	}

	var req walletdto.CreateTopupRequestRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	resp, err := h.service.CreateTopupRequest(r.Context(), uint(parsedUserID), req.Amount)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// SePayWebhook receives real-time bank transfer notifications from SePay.
//
// Processing flow (strict order):
//  1. Decode JSON body into typed struct (Code is *string for null safety)
//  2. Parse transaction date
//  3. INSERT into sepay_transactions with ON CONFLICT (id) DO NOTHING
//     — guarantees exactly-once processing per SePay transaction id
//  4. Return HTTP 200 with {"success": true} within milliseconds
//  5. Business logic (code matching, wallet crediting) runs in a goroutine
//
// Retry safety: If SePay retries with the same id, step 3 is a no-op
// and step 4 returns the same 200 response.
func (h *Handler) SePayWebhook(w http.ResponseWriter, r *http.Request) {
	var payload walletdto.SePayWebhookPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("[sepay] invalid webhook payload: %v", err)
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu webhook không hợp lệ")
		return
	}

	log.Printf("[sepay] received webhook: id=%d gateway=%s amount=%d transferType=%s content=%s",
		payload.ID, payload.Gateway, payload.TransferAmount, payload.TransferType, payload.Content)

	if err := h.service.HandleSePayWebhook(r.Context(), &payload); err != nil {
		log.Printf("[sepay] webhook processing error: id=%d err=%v", payload.ID, err)
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Exact required response format per SePay spec.
	// Any other status code or body shape will trigger retries.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
