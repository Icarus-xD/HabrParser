package repo

import (
	"github.com/Icarus-xD/HabrParser/internal/model"
	"gorm.io/gorm"
)

type LinkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) *LinkRepo {
	return &LinkRepo{
		db: db,
	}
}

func (r *LinkRepo) GetAll() ([]model.Link, error) {
	var links []model.Link

	err := r.db.Find(&links).Error
	if err != nil {
		return links, err
	}

	return links, nil
}