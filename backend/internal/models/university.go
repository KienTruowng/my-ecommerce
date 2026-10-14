package models

import "gorm.io/gorm"

type University struct {
	gorm.Model
	Name      string `gorm:"type:varchar(255);not null"`
	ShortName string `gorm:"type:varchar(50)"`

	Users []User `gorm:"foreignKey:UniversityID"`
}
