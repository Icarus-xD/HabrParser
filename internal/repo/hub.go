package repo

import (
	"errors"

	"github.com/Icarus-xD/HabrParser/internal/model"
	"gorm.io/gorm"
)

type HubRepo struct {
	db *gorm.DB
}

func NewHubRepo(db *gorm.DB) *HubRepo {
	return &HubRepo{
		db: db,
	}
}

func (r *HubRepo) Create(hubInfo model.Hub) (model.Hub, error) {
	var hub model.Hub 
	err := r.db.Where("link_id = ?", hubInfo.LinkID).First(&hub).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return hubInfo, err
	}

	if hub.LinkID > 0 {
		return hub, nil
	}

	err = r.db.Create(&hubInfo).Error
	if err != nil {
		return hubInfo, err
	}

	return hubInfo, nil
}