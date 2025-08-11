package main

import (
	"log"
	"os"
	"testing"

	"github.com/cheildo/deeli-api/internal/article"
	"github.com/cheildo/deeli-api/internal/auth"
	"github.com/cheildo/deeli-api/internal/recommendation"
	"github.com/cheildo/deeli-api/internal/user"
	"github.com/cheildo/deeli-api/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)

	if err := godotenv.Load("../../.env.test"); err != nil {
		log.Fatalf("Error loading .env.test file: %v", err)
	}

	database.Connect()
	clearTables() // Ensure tables are clean before migrations
	database.Migrate(&user.User{}, &article.Article{}, &article.Rating{})
	testRouter = setupRouter()

	exitCode := m.Run()

	clearTables()

	os.Exit(exitCode)
}

func setupRouter() *gin.Engine {
	// Repositories
	userRepo := user.NewRepository()
	articleRepo := article.NewRepository()

	// Services
	recommendationService := recommendation.NewService(articleRepo)

	// Handlers
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
		authRoutes.POST("/articles", articleHandler.CreateArticle)
		authRoutes.GET("/articles", articleHandler.GetArticles)
		authRoutes.GET("/recommendations", recommendationHandler.GetRecommendations)

	}

	return r
}

// clearTables removes all data from the tables to ensure a clean state for each test run.
func clearTables() {
	// The order matters due to foreign key constraints. Delete ratings/articles before users.
	database.DB.Exec("DELETE FROM ratings")
	database.DB.Exec("DELETE FROM articles")
	database.DB.Exec("DELETE FROM users")
}
