package models

import "gorm.io/gorm"

type ItemStatus string
type ItemType string

const (
	StatusAvailable ItemStatus = "available"
	StatusReserved  ItemStatus = "reserved"
	StatusTraded    ItemStatus = "traded"
	StatusHidden    ItemStatus = "hidden"

	TypeBarter  ItemType = "barter"
	TypeAuction ItemType = "auction"
)

type Item struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`

	Type   ItemType   `gorm:"type:varchar(20);default:'barter';not null"`
	Status ItemStatus `gorm:"type:varchar(20);default:'available';not null"`

	//---Relationship---//

	// Item belongs to User
	UserID uint
	User   User

	// Item has and belongs to many Categories
	Categories []*Category `gorm:"many2many:item_categories"`

	// Item might have many images
	ItemImage []*ItemImage

	Offers []*Offer `gorm:"foreignKey:TargetItemID"`
	Orders []*Order `gorm:"foreignKey:ItemID"`
}
