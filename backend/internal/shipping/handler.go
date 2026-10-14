package shipping

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/utils"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ListShippingServices godoc
// @Summary      List shipping services
// @Tags         Shipping
// @Produce      json
// @Success      200  {array}  models.ShippingService
// @Router       /api/shipping [get]
func (h *Handler) ListShipping(w http.ResponseWriter, r *http.Request) {
	services, err := h.repo.FindAll(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ")
		return
	}
	utils.WriteJSON(w, http.StatusOK, services)
}

// Admin List
func (h *Handler) AdminListShipping(w http.ResponseWriter, r *http.Request) {
	services, err := h.repo.FindAllAdmin(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi máy chủ")
		return
	}
	utils.WriteJSON(w, http.StatusOK, services)
}

type CreateShippingRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	BasePrice     uint    `json:"base_price"`
	PricePerKm    uint    `json:"price_per_km"`
	MinDistance   float64 `json:"min_distance"`
	EstimatedDays string  `json:"estimated_days"`
	IconURL       string  `json:"icon_url"`
}

func (h *Handler) AdminCreateShipping(w http.ResponseWriter, r *http.Request) {
	var req CreateShippingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "Tên dịch vụ không được để trống")
		return
	}
	s := &models.ShippingService{
		Name:          req.Name,
		Description:   req.Description,
		BasePrice:     req.BasePrice,
		PricePerKm:    req.PricePerKm,
		MinDistance:   req.MinDistance,
		EstimatedDays: req.EstimatedDays,
		IconURL:       req.IconURL,
		IsActive:      true,
	}
	if err := h.repo.Create(r.Context(), s); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi khi tạo")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, s)
}

func (h *Handler) AdminUpdateShipping(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	existing, err := h.repo.FindByID(r.Context(), uint(id))
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Không tìm thấy dịch vụ")
		return
	}

	var req CreateShippingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description
	existing.BasePrice = req.BasePrice
	existing.PricePerKm = req.PricePerKm
	existing.MinDistance = req.MinDistance
	existing.EstimatedDays = req.EstimatedDays
	existing.IconURL = req.IconURL

	if err := h.repo.Update(r.Context(), existing); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi khi cập nhật")
		return
	}
	utils.WriteJSON(w, http.StatusOK, existing)
}

func (h *Handler) AdminDeleteShipping(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}
	if err := h.repo.Delete(r.Context(), uint(id)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Lỗi khi xóa")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Đã xóa"})
}

// ─── Offer-type suggestion endpoint ───────────────────────────────────────

type OfferTypeResponse struct {
	Type               string                 `json:"type"` // "meetup" | "campus_trade" | "shipping"
	DistanceKm         float64                `json:"distance_km"`
	Message            string                 `json:"message"`
	SuggestedLocations []models.Location      `json:"suggested_locations,omitempty"`
	ShippingOptions    []ShippingCostEstimate `json:"shipping_options,omitempty"`
}

type ShippingCostEstimate struct {
	Service models.ShippingService `json:"service"`
	Cost    uint                   `json:"cost"`
}

// GetOfferType determines what kind of offer flow to show based on:
// - Distance between buyer & seller
// - Whether an item is being offered (trade) vs cash only
func (h *Handler) GetOfferType(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	buyerLatStr := r.URL.Query().Get("buyer_lat")
	buyerLngStr := r.URL.Query().Get("buyer_lng")
	sellerLatStr := r.URL.Query().Get("seller_lat")
	sellerLngStr := r.URL.Query().Get("seller_lng")
	hasItemStr := r.URL.Query().Get("has_item") // "true" if trading an item

	buyerLat, _ := strconv.ParseFloat(buyerLatStr, 64)
	buyerLng, _ := strconv.ParseFloat(buyerLngStr, 64)
	sellerLat, _ := strconv.ParseFloat(sellerLatStr, 64)
	sellerLng, _ := strconv.ParseFloat(sellerLngStr, 64)
	hasItem := hasItemStr == "true"

	resp := OfferTypeResponse{}

	// If either user has no coordinates, default to meetup
	if buyerLat == 0 || sellerLat == 0 {
		resp.Type = "meetup"
		resp.Message = "Could not determine distance. Please set your address in profile settings."
		utils.WriteJSON(w, http.StatusOK, resp)
		return
	}

	distance := HaversineDistance(buyerLat, buyerLng, sellerLat, sellerLng)
	resp.DistanceKm = distance

	const CAMPUS_RADIUS = 10.0 // 10km

	if distance <= CAMPUS_RADIUS {
		if hasItem {
			// Campus trade: AI picks from pre-inserted locations
			resp.Type = "campus_trade"
			resp.Message = "Both users are within campus range. Meet at a safe campus location to exchange items."

			// Get midpoint and find nearby locations
			midLat, midLng := Midpoint(buyerLat, buyerLng, sellerLat, sellerLng)
			var locations []models.Location
			h.repo.db.WithContext(r.Context()).Where("is_active = ?", true).Find(&locations)
			resp.SuggestedLocations = PickMeetupLocations(midLat, midLng, locations, 3)
		} else {
			// Cash only within 10km → make appointment, AI picks location
			resp.Type = "meetup"
			resp.Message = "You're close by! Meet up at a convenient location to complete the transaction."

			midLat, midLng := Midpoint(buyerLat, buyerLng, sellerLat, sellerLng)
			var locations []models.Location
			h.repo.db.WithContext(r.Context()).Where("is_active = ?", true).Find(&locations)
			resp.SuggestedLocations = PickMeetupLocations(midLat, midLng, locations, 3)
		}
	} else {
		// > 10km → shipping (cash only, item trade not possible)
		if hasItem {
			resp.Type = "too_far"
			resp.Message = "Users are too far apart for item exchange. Only cash offers with shipping are available when distance exceeds 10km."
		} else {
			resp.Type = "shipping"
			resp.Message = "You're far apart. Use a shipping service to complete the transaction."

			services, err := h.repo.FindAll(r.Context())
			if err == nil {
				for _, svc := range services {
					if svc.MinDistance <= distance {
						cost := CalculateShippingCost(svc, distance)
						resp.ShippingOptions = append(resp.ShippingOptions, ShippingCostEstimate{
							Service: svc,
							Cost:    cost,
						})
					}
				}
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
