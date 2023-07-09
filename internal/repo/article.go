package repo

import (
	"errors"

	"github.com/Icarus-xD/HabrParser/internal/model"
	"gorm.io/gorm"
)

type ArticleRepo struct {
	db *gorm.DB
}

func NewArticleRepo(db *gorm.DB) *ArticleRepo {
	return &ArticleRepo{
		db: db,
	}
}

func (r *ArticleRepo) Create(articleInfo model.Article) error {
	var article model.Article
	err := r.db.Where("url = ?", articleInfo.URL).First(&article).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if article.HubID > 0 {
		return nil
	}

	err = r.db.Create(&articleInfo).Error
	if err != nil {
		return err
	}

	return nil
}