package model

import (
	"gorm.io/gorm"
)

type Hub struct {
	gorm.Model
	LinkID      uint
	Link      	Link    `gorm:"foreignKey:LinkID"`
	Title       string  `gorm:"not null"`
	Description string  `gorm:"not null"`
	Rating      float64 `gorm:"not null"`
}

func (Hub) TableName() string {
	return "hubs"
}