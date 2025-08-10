package database

import (
	"log"

	"github.com/cheildo/deeli-api/internal/article"
	"github.com/cheildo/deeli-api/internal/user"
	"github.com/cheildo/deeli-api/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := config.Get("DB_SOURCE")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connection successful.")
}

func Migrate() {
	err := DB.AutoMigrate(&user.User{}, &article.Article{}, &article.Rating{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration successful.")
}
