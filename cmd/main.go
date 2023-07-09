package main

import (
	"fmt"
	"log"

	"github.com/Icarus-xD/HabrParser/internal/config"
	"github.com/Icarus-xD/HabrParser/internal/database"
	"github.com/Icarus-xD/HabrParser/internal/repo"
	"github.com/Icarus-xD/HabrParser/internal/service"
)


func main() {
	// Config
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// DB
	dsn := fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PostgresDriver, cfg.PostgresUser, cfg.PostgresPassword,
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB, cfg.PostgresSSLMode,
	)

	db, err := database.Init(dsn)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Repo
	linkRepo := repo.NewLinkRepo(db)
	hubRepo := repo.NewHubRepo(db)
	articleRepo := repo.NewArticleRepo(db)

	// Service
	crawler := service.NewCrawler(linkRepo, hubRepo, articleRepo)
	err = crawler.RunCrawling()
	if err != nil {
		log.Fatal(err)
	}
}