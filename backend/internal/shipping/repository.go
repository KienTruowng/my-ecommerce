package shipping

import (
	"context"
	"math"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAll(ctx context.Context) ([]models.ShippingService, error) {
	var services []models.ShippingService
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&services).Error
	return services, err
}

func (r *Repository) FindAllAdmin(ctx context.Context) ([]models.ShippingService, error) {
	var services []models.ShippingService
	err := r.db.WithContext(ctx).Order("id ASC").Find(&services).Error
	return services, err
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*models.ShippingService, error) {
	var s models.ShippingService
	err := r.db.WithContext(ctx).First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Create(ctx context.Context, s *models.ShippingService) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) Update(ctx context.Context, s *models.ShippingService) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ShippingService{}, id).Error
}

// ─── Distance & Location helpers ───────────────────────────────────────────

// HaversineDistance returns kilometers between two lat/lng points.
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// Midpoint returns the geographic midpoint of two lat/lng points.
func Midpoint(lat1, lng1, lat2, lng2 float64) (float64, float64) {
	lat1r := lat1 * math.Pi / 180
	lng1r := lng1 * math.Pi / 180
	lat2r := lat2 * math.Pi / 180
	lng2r := lng2 * math.Pi / 180

	bx := math.Cos(lat2r) * math.Cos(lng2r-lng1r)
	by := math.Cos(lat2r) * math.Sin(lng2r-lng1r)
	midLat := math.Atan2(math.Sin(lat1r)+math.Sin(lat2r),
		math.Sqrt((math.Cos(lat1r)+bx)*(math.Cos(lat1r)+bx)+by*by))
	midLng := lng1r + math.Atan2(by, math.Cos(lat1r)+bx)

	return midLat * 180 / math.Pi, midLng * 180 / math.Pi
}

// CalculateShippingCost returns the estimated cost for a given distance.
func CalculateShippingCost(s models.ShippingService, distanceKm float64) uint {
	cost := s.BasePrice
	if distanceKm > 0 {
		cost += uint(math.Ceil(distanceKm)) * s.PricePerKm
	}
	return cost
}

// ─── Location Picker ──────────────────────────────────────────────────────

// PickMeetupLocations finds the best locations near the midpoint of two users.
// Scores by: distance from midpoint, open hours, popularity.
func PickMeetupLocations(
	midLat, midLng float64,
	locations []models.Location,
	topN int,
) []models.Location {
	if topN <= 0 {
		topN = 3
	}

	type scored struct {
		loc   models.Location
		score float64
	}

	var scoredLocs []scored
	for _, loc := range locations {
		if !loc.IsActive || loc.Latitude == nil || loc.Longitude == nil {
			continue
		}
		dist := HaversineDistance(midLat, midLng, *loc.Latitude, *loc.Longitude)
		// Score: closer = better (invert distance)
		score := 100.0 / (1.0 + dist)
		// Bonus for well-documented locations
		if len(loc.Description) > 20 {
			score += 10
		}
		scoredLocs = append(scoredLocs, scored{loc, score})
	}

	// Sort descending by score
	for i := 0; i < len(scoredLocs); i++ {
		for j := i + 1; j < len(scoredLocs); j++ {
			if scoredLocs[j].score > scoredLocs[i].score {
				scoredLocs[i], scoredLocs[j] = scoredLocs[j], scoredLocs[i]
			}
		}
	}

	result := make([]models.Location, 0, topN)
	for i := 0; i < topN && i < len(scoredLocs); i++ {
		result = append(result, scoredLocs[i].loc)
	}
	return result
}
