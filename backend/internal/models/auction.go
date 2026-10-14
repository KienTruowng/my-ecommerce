package models

import (
	"time"

	"gorm.io/gorm"
)

type AuctionStatus string
type BidStatus string

const (
	AuctionStatusOpen      AuctionStatus = "open"
	AuctionStatusSettled   AuctionStatus = "settled"
	AuctionStatusCancelled AuctionStatus = "cancelled"

	BidStatusActive    BidStatus = "active"
	BidStatusWon       BidStatus = "won"
	BidStatusLost      BidStatus = "lost"
	BidStatusForfeited BidStatus = "forfeited" // winner didn't complete → deposit forfeited to runner-up
)

// AuctionConfig holds the seller-defined parameters for an auction item.
// An item of Type == TypeAuction must have exactly one AuctionConfig.
//
// Business rules:
//   - StepPrice must be ≤ 20% of BasePrice (enforced at service layer)
//   - DepositAmount is the security deposit a buyer must "pay" before bidding
//   - Once deadline passes the auction is settled by the admin/cron endpoint
type AuctionConfig struct {
	gorm.Model

	// FK → items.id  (unique: one config per auction item)
	ItemID uint `gorm:"uniqueIndex;not null" json:"item_id"`
	Item   Item `gorm:"foreignKey:ItemID"   json:"item,omitempty"`

	BasePrice     uint      `gorm:"not null"           json:"base_price"`
	StepPrice     uint      `gorm:"not null"           json:"step_price"`
	DepositAmount uint      `gorm:"not null"           json:"deposit_amount"`
	Deadline      time.Time `gorm:"not null"           json:"deadline"`

	Status AuctionStatus `gorm:"type:varchar(20);default:'open';not null" json:"status"`

	Bids []*Bid `gorm:"foreignKey:AuctionConfigID" json:"bids,omitempty"`
}

// Bid represents a single bid placed by a buyer in an auction.
// Rank is recalculated (1 = highest) each time a new bid is placed.
type Bid struct {
	gorm.Model

	AuctionConfigID uint          `gorm:"not null;index"                          json:"auction_config_id"`
	BidderID        uint          `gorm:"not null;index"                          json:"bidder_id"`
	Amount          uint          `gorm:"not null"                                json:"amount"`
	DepositPaid     bool          `gorm:"default:false"                           json:"deposit_paid"` // placeholder — true = deposit reserved
	Status          BidStatus     `gorm:"type:varchar(20);default:'active';not null" json:"status"`
	Rank            int           `gorm:"default:0"                               json:"rank"` // 1 = current highest

	Bidder        User          `gorm:"foreignKey:BidderID"        json:"bidder,omitempty"`
	AuctionConfig AuctionConfig `gorm:"foreignKey:AuctionConfigID" json:"auction_config,omitempty"`
}
