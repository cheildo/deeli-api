package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/cheildo/deeli-api/internal/recommendation"

	"github.com/cheildo/deeli-api/internal/article"
	"github.com/cheildo/deeli-api/internal/auth"
	"github.com/cheildo/deeli-api/internal/user"
	"github.com/cheildo/deeli-api/internal/worker"
	"github.com/cheildo/deeli-api/pkg/config"
	"github.com/cheildo/deeli-api/pkg/database"
)

func main() {
	config.LoadConfig()
	database.Connect()
	database.Migrate(&user.User{}, &article.Article{}, &article.Rating{})

	// --- Repositories ---
	userRepo := user.NewRepository()
	articleRepo := article.NewRepository()

	// --- Services ---
	recommendationService := recommendation.NewService(articleRepo)

	// --- Start Background Worker ---
	bgWorker := worker.NewWorker(articleRepo)
	go bgWorker.Start()

	// --- Handlers ---
	userHandler := user.NewHandler(userRepo)
	articleHandler := article.NewHandler(articleRepo)

	recommendationHandler := recommendation.NewHandler(recommendationService)

	r := gin.Default()

	// Public routes
	r.POST("/signup", userHandler.Signup)
	r.POST("/login", userHandler.Login)

	// Authenticated routes
	authRoutes := r.Group("/")
	authRoutes.Use(auth.Middleware())
	{
		authRoutes.GET("/me", userHandler.GetMe)

		// Article routes
		authRoutes.POST("/articles", articleHandler.CreateArticle)
		authRoutes.GET("/articles", articleHandler.GetArticles)
		authRoutes.DELETE("/articles/:id", articleHandler.DeleteArticle)

		// Rating routes
		authRoutes.POST("/articles/:id/rate", articleHandler.RateArticle)
		authRoutes.GET("/articles/:id/rate", articleHandler.GetRating)
		authRoutes.DELETE("/articles/:id/rate", articleHandler.DeleteRating)

		// Recommendation route
		authRoutes.GET("/recommendations", recommendationHandler.GetRecommendations)
	}

	// Start the server
	serverAddress := config.Get("SERVER_ADDRESS")
	log.Printf("Starting server on %s", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
