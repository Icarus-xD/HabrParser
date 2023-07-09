package database

import (
	"errors"

	"github.com/Icarus-xD/HabrParser/internal/model"
	"gorm.io/gorm"
)

func runMigration(db *gorm.DB) error {
	hubUrls := []string{
		"https://habr.com/ru/hub/popular_science/",
		"https://habr.com/ru/hub/itcompanies/",
		"https://habr.com/ru/hub/programming/",
	}

	var err error

	// // Link
	if !db.Migrator().HasTable(&model.Link{}) {
		err = db.AutoMigrate(&model.Link{})
	}
	if err != nil {
		return err
	}

	for _, url := range hubUrls {
		var link model.Link
		err := db.Where("url = ?", url).First(&link).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if len(link.URL) != 0 {
			return nil
		}

		link = model.Link{
			URL: url,
		}

		_ = db.Create(&link).Error
	}

	// Hub
	if !db.Migrator().HasTable(&model.Hub{}) {
		err = db.AutoMigrate(&model.Hub{})
	}
	if err != nil {
		return err
	}

	// Article
	if !db.Migrator().HasTable(&model.Article{}) {
		err = db.AutoMigrate(&model.Article{})
	}
	if err != nil {
		return err
	}

	return nil
}