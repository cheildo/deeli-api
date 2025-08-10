package main

import (
	"log"

	"github.com/cheildo/deeli-api/internal/article"
	"github.com/cheildo/deeli-api/internal/auth"
	"github.com/cheildo/deeli-api/internal/user"
	"github.com/cheildo/deeli-api/pkg/config"
	"github.com/cheildo/deeli-api/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	database.Connect()
	database.Migrate(&user.User{}, &article.Article{}, &article.Rating{})

	r := gin.Default()

	// Repositories
	userRepo := user.NewRepository()

	// Handlers
	userHandler := user.NewHandler(userRepo)

	// Public routes
	r.POST("/signup", userHandler.Signup)
	r.POST("/login", userHandler.Login)

	// Authenticated routes
	authRoutes := r.Group("/")
	authRoutes.Use(auth.Middleware())
	{
		authRoutes.GET("/me", userHandler.GetMe)
	}

	// Start the server
	serverAddress := config.Get("SERVER_ADDRESS")
	log.Printf("Starting server on %s", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
