package models

import "gorm.io/gorm"

type Location struct {
	gorm.Model
	Name        string   `gorm:"type:varchar(255);not null"`
	Description string   `gorm:"type:text"`
	Address     string   `gorm:"type:varchar(255)"`
	ImageURL    string   `gorm:"type:text"`
	Latitude    *float64 `gorm:"type:decimal(10,7)"`
	Longitude   *float64 `gorm:"type:decimal(10,7)"`
	OpenTime    string   `gorm:"type:varchar(10)"`
	CloseTime   string   `gorm:"type:varchar(10)"`
	IsActive    bool     `gorm:"default:true"`
}
