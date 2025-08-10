package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads config from .env file
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func Get(key string) string {
	return os.Getenv(key)
}
