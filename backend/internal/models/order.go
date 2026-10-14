package models

import "gorm.io/gorm"

type OrderStatus string
type PaymentStatus string

const (
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"

	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type Order struct {
	gorm.Model

	BuyerID       uint          `gorm:"not null;index"`
	SellerID      uint          `gorm:"not null;index"`
	ItemID        uint          `gorm:"not null;index"`
	TradeInItemID *uint         `gorm:"index"`
	CashAmount    uint          `gorm:"default:0"`
	PaymentMethod string        `gorm:"type:varchar(30);default:'wallet'"`
	PaymentStatus PaymentStatus `gorm:"type:varchar(20);default:'pending'"`
	OrderStatus   OrderStatus   `gorm:"type:varchar(20);default:'confirmed'"`

	Buyer       User         `gorm:"foreignKey:BuyerID"`
	Seller      User         `gorm:"foreignKey:SellerID"`
	Item        Item         `gorm:"foreignKey:ItemID"`
	TradeInItem *Item        `gorm:"foreignKey:TradeInItemID"`
	Appointment *Appointment `gorm:"foreignKey:OrderID"`
}
