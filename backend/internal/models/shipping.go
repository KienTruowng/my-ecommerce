package models

import "gorm.io/gorm"

type ShippingService struct {
	gorm.Model
	Name          string  `gorm:"type:varchar(255);not null" json:"name"`
	Description   string  `gorm:"type:text" json:"description"`
	BasePrice     uint    `gorm:"default:0" json:"base_price"`            // base fee (VND)
	PricePerKm    uint    `gorm:"default:0" json:"price_per_km"`          // additional per km
	MinDistance   float64 `gorm:"default:0" json:"min_distance"`          // minimum km to offer shipping
	EstimatedDays string  `gorm:"type:varchar(50)" json:"estimated_days"` // e.g. "1-3"
	IconURL       string  `gorm:"type:text" json:"icon_url"`
	IsActive      bool    `gorm:"default:true" json:"is_active"`
}
