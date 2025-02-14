package e2e_tests

import (
	"avito_staj_2025/domain"
	auth "avito_staj_2025/internal/auth/controller"
	authRepository "avito_staj_2025/internal/auth/repository"
	authUsecase "avito_staj_2025/internal/auth/usecase"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres dbname=test password=qaleka123 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&domain.User{})
	assert.NoError(t, err)

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	err := db.Migrator().DropTable(&domain.User{})
	assert.NoError(t, err)
}

func createTestUser(t *testing.T, db *gorm.DB, username, password string) {
	hashedPassword, err := middleware.HashPassword(password)
	assert.NoError(t, err)

	user := domain.User{
		Username: username,
		Password: hashedPassword,
		Coins:    1000,
	}
	err = db.Create(&user).Error
	assert.NoError(t, err)
}

func TestLoginUserE2E(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	jwtToken, err := middleware.NewJwtToken("secret-key")
	assert.NoError(t, err)

	err = logger.InitLoggers()
	assert.NoError(t, err)
	defer func() {
		err := logger.SyncLoggers()
		assert.NoError(t, err)
	}()

	username := "test_user"
	password := "test_password"
	createTestUser(t, db, username, password)

	authRepo := authRepository.NewAuthRepository(db)
	authUC := authUsecase.NewAuthUsecase(authRepo)
	authHandler := auth.NewAuthHandler(authUC, jwtToken)

	router := mux.NewRouter()
	api := "/api"

	router.HandleFunc(api+"/auth", authHandler.LoginUser).Methods("POST")

	server := httptest.NewServer(router)
	defer server.Close()

	requestBody := domain.LoginRequest{
		Username: username,
		Password: password,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/auth", bytes.NewBuffer(requestBodyBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	token, exists := response["token"]
	assert.True(t, exists)
	assert.NotEmpty(t, token)

	_, err = jwtToken.Validate(token.(string))
	assert.NoError(t, err)
}

func TestLoginUserFirstTimeE2E(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	jwtToken, err := middleware.NewJwtToken("secret-key")
	assert.NoError(t, err)

	err = logger.InitLoggers()
	assert.NoError(t, err)
	defer func() {
		err := logger.SyncLoggers()
		assert.NoError(t, err)
	}()

	username := "test_user"
	password := "test_password"

	authRepo := authRepository.NewAuthRepository(db)
	authUC := authUsecase.NewAuthUsecase(authRepo)
	authHandler := auth.NewAuthHandler(authUC, jwtToken)

	router := mux.NewRouter()
	api := "/api"

	router.HandleFunc(api+"/auth", authHandler.LoginUser).Methods("POST")

	server := httptest.NewServer(router)
	defer server.Close()

	requestBody := domain.LoginRequest{
		Username: username,
		Password: password,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/auth", bytes.NewBuffer(requestBodyBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	token, exists := response["token"]
	assert.True(t, exists)
	assert.NotEmpty(t, token)

	_, err = jwtToken.Validate(token.(string))
	assert.NoError(t, err)
}
