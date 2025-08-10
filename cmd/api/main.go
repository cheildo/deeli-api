package main

import (
	"log"

	"github.com/cheildo/deeli-api/pkg/config"
	"github.com/cheildo/deeli-api/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	database.Connect()
	database.Migrate()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Start the server
	serverAddress := config.Get("SERVER_ADDRESS")
	log.Printf("Starting server on %s", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
