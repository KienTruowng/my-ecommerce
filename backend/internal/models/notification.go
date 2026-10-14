package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	UserID  uint   `gorm:"not null;index" json:"user_id"`
	Title   string `gorm:"type:varchar(255);not null" json:"title"`
	Message string `gorm:"type:text" json:"message"`
	Link    string `gorm:"type:varchar(255)" json:"link"` // e.g. "/offers"
	IsRead  bool   `gorm:"default:false" json:"is_read"`
	Type    string `gorm:"type:varchar(50)" json:"type"` // "offer_received", "offer_accepted", "offer_rejected"
}
