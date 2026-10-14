package aichat

import (
	"encoding/json"
	"net/http"

	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		utils.WriteError(w, http.StatusServiceUnavailable, "AI service not configured. Set DEEPSEEK_API_KEY in .env")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Messages required")
		return
	}

	resp, err := h.service.Chat(req.Messages)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "AI error: "+err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
