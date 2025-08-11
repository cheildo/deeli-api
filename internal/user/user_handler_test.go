package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cheildo/deeli-api/internal/auth"
	"github.com/cheildo/deeli-api/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

// setup performs the setup for the tests in this file.
func setup() {
	gin.SetMode(gin.TestMode)

	if err := godotenv.Load("../../.env.test"); err != nil {
		panic("Error loading .env.test file for user test: " + err.Error())
	}

	database.Connect()
	// We use our User model here directly for migration.
	database.DB.AutoMigrate(&User{})

	userRepo := NewRepository()
	userHandler := NewHandler(userRepo)

	router = gin.Default()
	router.POST("/signup", userHandler.Signup)
	router.POST("/login", userHandler.Login)
	authRoutes := router.Group("/")
	authRoutes.Use(auth.Middleware())
	{
		authRoutes.GET("/me", userHandler.GetMe)
	}
}

// teardown cleans up after tests.
func teardown() {
	database.DB.Exec("DELETE FROM users")
}

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}

func TestSignupAndLoginFlow(t *testing.T) {
	teardown()
	defer teardown()

	// 1. Test Successful Signup
	signupPayload := `{"email": "flow@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBufferString(signupPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Test Duplicate Signup
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/signup", bytes.NewBufferString(signupPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	// 3. Test Login with wrong password
	loginPayloadWrong := `{"email": "flow@example.com", "password": "wrongpassword"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginPayloadWrong))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 4. Test Login with correct password
	loginPayloadCorrect := `{"email": "flow@example.com", "password": "password123"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginPayloadCorrect))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	token, exists := response["token"]
	assert.True(t, exists)
	assert.NotEmpty(t, token)

	// 5. Test /me endpoint with the token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var meResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &meResponse)
	assert.NoError(t, err)
	assert.Equal(t, "flow@example.com", meResponse["email"])
}
