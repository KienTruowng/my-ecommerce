package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	University   *University `json:"university,omitempty"`
	UniversityID *uint       `json:"university_id"`
	Name         string      `json:"Name"`
	Email        string      `gorm:"uniqueIndex;not null" json:"email"`
	Password     string      `gorm:"not null" json:"-"`
	AvatarURL    string      `json:"avatarUrl"`

	// Address fields — full address is private, only city shown to others
	City      string   `json:"-"`
	District  string   `json:"-"`
	Latitude  *float64 `gorm:"type:decimal(10,7)" json:"-"`
	Longitude *float64 `gorm:"type:decimal(10,7)" json:"-"`

	IsVerified bool   `gorm:"default:false"`
	Role       string `gorm:"default:'user'" json:"Role"`
}

// PublicCity returns only the city name for public display.
func (u *User) PublicCity() string { return u.City }
