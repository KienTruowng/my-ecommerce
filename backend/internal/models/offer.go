package models

import "gorm.io/gorm"

type OfferStatus string

const (
	OfferStatusPending   OfferStatus = "pending"
	OfferStatusAccepted  OfferStatus = "accepted"
	OfferStatusRejected  OfferStatus = "rejected"
	OfferStatusCancelled OfferStatus = "cancelled"
)

// Current logic
// Offer represents a transaction proposal in the Barter flow.
// The current barter logic supports 3 flexible scenarios for students:
//
//  1. Pure Barter (Item for Item): The user selects 1 of their items to exchange for another item.
//     => OfferItemID != nil, CashAmount = 0
//
//  2. Barter + Cash Top-up: The user offers an item of lower value and adds cash to make up the difference.
//     => OfferItemID != nil, CashAmount > 0
//
//  3. Cash Only (Buyout): The user has no item to trade and only wants to buy the item with cash.
//     => OfferItemID == nil, CashAmount > 0
//
// Lifecycle:
//   - Upon creation, an Offer defaults to the 'pending' status.
//   - The owner of the TargetItem has the right to review and select 'accepted' or 'rejected'.
//   - Typically, when 1 Offer is 'accepted', the system at the Service layer will automatically
//     trigger the transition of all remaining 'pending' Offers for that TargetItemID to 'rejected'.
type Offer struct {
	gorm.Model

	OfferUserID uint `gorm:"not null;index"`

	// TargetItemID: The item the student WANTS TO ACQUIRE (Always required)
	TargetItemID uint  `gorm:"not null;index" json:"target_item_id"`
	OfferItemID  *uint `gorm:"index" json:"offer_item_id,omitempty"`
	CashAmount   uint  `gorm:"default:0" json:"cash_amount"`

	OfferMessage string `gorm:"type:varchar(500)"`

	PaymentMethod string `gorm:"type:varchar(30);default:'cod'" json:"payment_method"` // "cod" or "account"

	Status OfferStatus `gorm:"type:varchar(20);default:'pending';not null"`

	User        User         `gorm:"foreignKey:OfferUserID" json:"user,omitempty"`
	TargetItem  Item         `gorm:"foreignKey:TargetItemID" json:"target_item,omitempty"`
	OfferItem   *Item        `gorm:"foreignKey:OfferItemID" json:"offer_item,omitempty"`
	Appointment *Appointment `gorm:"foreignKey:OfferID" json:"appointment,omitempty"`
}
