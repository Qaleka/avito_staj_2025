package e2etests

import (
	"avito_staj_2025/domain"
	auth "avito_staj_2025/internal/auth/controller"
	authRepository "avito_staj_2025/internal/auth/repository"
	authUsecase "avito_staj_2025/internal/auth/usecase"
	dsn2 "avito_staj_2025/internal/service/dsn"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func createDatabaseIfNotExists() error {
	host := os.Getenv("DB_HOST_TEST")
	port := os.Getenv("DB_PORT_TEST")
	user := os.Getenv("DB_USER_TEST")
	pass := os.Getenv("DB_PASS_TEST")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", host, port, user, pass)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	var count int64
	db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = 'test'").Scan(&count)

	if count == 0 {
		_ = db.Exec("CREATE DATABASE test").Error
	}

	sqlDB, _ := db.DB()
	sqlDB.Close()

	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	err := createDatabaseIfNotExists()
	assert.NoError(t, err)
	dsn := dsn2.FromEnvE2E()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err = tx.AutoMigrate(&domain.User{})
	assert.NoError(t, err)
	tx.Commit()
	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	err := db.Migrator().DropTable(&domain.User{})
	assert.NoError(t, err)
}

func createTestUser(t *testing.T, db *gorm.DB, username, password string) {

	user := domain.User{
		Username: username,
		Password: password,
		Coins:    1000,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)
}

func TestLoginUserE2E(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	db := setupTestDB(t)
	jwtToken, err := middleware.NewJwtToken("secret-key")
	assert.NoError(t, err)

	err = logger.InitLoggers()
	assert.NoError(t, err)
	defer func() {
		err := logger.SyncLoggers()
		assert.NoError(t, err)
	}()

	username := fmt.Sprintf("user_%d", time.Now().Unix())
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
	_ = godotenv.Load("../../../.env")
	db := setupTestDB(t)
	jwtToken, err := middleware.NewJwtToken("secret-key")
	assert.NoError(t, err)

	err = logger.InitLoggers()
	assert.NoError(t, err)
	defer func() {
		err := logger.SyncLoggers()
		assert.NoError(t, err)
	}()

	username := fmt.Sprintf("user_%d", time.Now().Unix())
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
