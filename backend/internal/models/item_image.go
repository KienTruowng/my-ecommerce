package models

import "gorm.io/gorm"

type ItemImage struct {
	gorm.Model
	URL    string
	ItemID uint
}
