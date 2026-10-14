package appointment

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

func (h *Handler) GetMyAppointments(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	appts, err := h.service.GetMyAppointments(r.Context(), uint(userID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ nội bộ")
		return
	}

	resp := make([]AppointmentResponse, 0, len(appts))
	for _, a := range appts {
		partnerName := ""
		if a.Order != nil {
			if a.Order.BuyerID == uint(userID) {
				partnerName = a.Order.Seller.Name
			} else {
				partnerName = a.Order.Buyer.Name
			}
		} else if a.Offer != nil {
			if a.Offer.OfferUserID == uint(userID) {
				partnerName = a.Offer.TargetItem.User.Name
			} else {
				partnerName = a.Offer.User.Name
			}
		}

		resp = append(resp, AppointmentResponse{
			ID:          a.ID,
			OfferID:     a.OfferID,
			OrderID:     a.OrderID,
			Title:       a.Title,
			Date:        a.Date,
			Time:        a.Time,
			Location:    a.Location,
			Note:        a.Note,
			Status:      string(a.Status),
			PartnerName: partnerName,
			CreatedAt:   a.CreatedAt,
		})
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := appctx.GetUserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	apptIDStr := r.PathValue("id")
	apptID, err := strconv.ParseUint(apptIDStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid appointment ID")
		return
	}

	var req UpdateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	appt, err := h.service.UpdateAppointment(r.Context(), uint(userID), uint(apptID), &req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrForbidden) {
			utils.WriteError(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, appt)
}
