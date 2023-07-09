package model

import (
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	URL       string `gorm:"unique;not null"`
}

func (Link) TableName() string {
	return "links"
}