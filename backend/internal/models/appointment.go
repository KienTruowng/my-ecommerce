package models

import "gorm.io/gorm"

type AppointmentStatus string

const (
	ApptStatusScheduled AppointmentStatus = "scheduled"
	ApptStatusConfirmed AppointmentStatus = "confirmed"
	ApptStatusCompleted AppointmentStatus = "completed"
	ApptStatusCancelled AppointmentStatus = "cancelled"
)

type Appointment struct {
	gorm.Model
	OfferID *uint `gorm:"index"`
	Offer   *Offer `gorm:"foreignKey:OfferID"`
	OrderID *uint `gorm:"index"`
	Order   *Order `gorm:"foreignKey:OrderID"`

	Title    string            `gorm:"type:varchar(255)"`
	Date     string            `gorm:"type:varchar(20)"`
	Time     string            `gorm:"type:varchar(20)"`
	Location string            `gorm:"type:varchar(255)"`
	Note     string            `gorm:"type:text"`
	Status   AppointmentStatus `gorm:"type:varchar(20);default:'scheduled'"`
}
