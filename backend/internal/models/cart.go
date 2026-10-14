package models

import "gorm.io/gorm"

type Cart struct {
	gorm.Model

	UserID        uint  `gorm:"uniqueIndex;not null" json:"user_id"`
	ItemID        uint  `gorm:"not null;index" json:"item_id"`
	TradeInItemID *uint `gorm:"index" json:"offer_item_id,omitempty"`
	CashAmount    uint  `gorm:"default:0" json:"cash_amount"`

	User        User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Item        *Item  `gorm:"foreignKey:ItemID" json:"item"`
	TradeInItem *Item  `gorm:"foreignKey:TradeInItemID" json:"trade_in_item,omitempty"`
}
