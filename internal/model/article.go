package model

import (
	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	HubID      uint
	Hub        Hub    `gorm:"foreignKey:HubID"`
	URL        string `gorm:"unique;not null"`
	Title      string `gorm:"not null"`
	Datetime   string `gorm:"not null"`
	Author     string 
	AuthorLink string 
}

func (Article) TableName() string {
	return "articles"
}