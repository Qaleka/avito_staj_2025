package e2etests

import (
	"avito_staj_2025/domain"
	auth "avito_staj_2025/internal/auth/controller"
	authRepository "avito_staj_2025/internal/auth/repository"
	authUsecase "avito_staj_2025/internal/auth/usecase"
	merchController "avito_staj_2025/internal/merch/controller"
	merchRepository "avito_staj_2025/internal/merch/repository"
	merchUsecase "avito_staj_2025/internal/merch/usecase"
	"avito_staj_2025/internal/service/dsn"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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
	dsn := dsn.FromEnvE2E()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = tx.AutoMigrate(&domain.User{}, &domain.Inventory{}, &domain.Transaction{})
	assert.NoError(t, err)

	tx.Commit()
	assert.NoError(t, err)

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	err := db.Migrator().DropTable(&domain.User{}, &domain.Inventory{}, &domain.Transaction{})
	assert.NoError(t, err)
}

func createTestUser(t *testing.T, db *gorm.DB, userID, username string, coins int) {
	user := domain.User{
		UUID:     userID,
		Username: username,
		Coins:    coins,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)
}

func createTestInventory(t *testing.T, db *gorm.DB, userID, itemName string, itemAmount int) {
	inventory := domain.Inventory{
		OwnerID:    userID,
		ItemName:   itemName,
		ItemAmount: itemAmount,
	}
	err := db.Create(&inventory).Error
	assert.NoError(t, err)
}

func createTestTransaction(t *testing.T, db *gorm.DB, senderID, receiverID string, amount int) {
	transaction := domain.Transaction{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     amount,
	}
	err := db.Create(&transaction).Error
	assert.NoError(t, err)
}

func TestBuyItemE2E(t *testing.T) {
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

	userID := uuid.New().String()
	username := fmt.Sprintf("u_%d", time.Now().UnixNano())
	createTestUser(t, db, userID, username, 500)

	token, err := jwtToken.Create(userID, time.Now().Add(24*time.Hour).Unix())
	assert.NoError(t, err)

	itemName := "pen"
	domain.MerchTypes = map[string]int{
		itemName: 100,
	}

	authRepo := authRepository.NewAuthRepository(db)
	authUC := authUsecase.NewAuthUsecase(authRepo)
	authHandler := auth.NewAuthHandler(authUC, jwtToken)

	merchRepo := merchRepository.NewMerchRepository(db)
	merchUC := merchUsecase.NewMerchUsecase(merchRepo)
	merchHandler := merchController.NewMerchHandler(merchUC, jwtToken)

	router := mux.NewRouter()
	api := "/api"

	router.HandleFunc(api+"/auth", authHandler.LoginUser).Methods("POST")
	router.HandleFunc(api+"/info", merchHandler.GetUserMerchInformation).Methods("GET")
	router.HandleFunc(api+"/buy/{item}", merchHandler.BuyItem).Methods("GET")
	router.HandleFunc(api+"/sendCoin", merchHandler.SendCoins).Methods("POST")

	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", server.URL, itemName), nil)
	assert.NoError(t, err)
	req.Header.Set("JWT-Token", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var user domain.User
	err = db.Where("uuid = ?", userID).First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, 400, user.Coins) // 500 - 100 = 400

	var inventory domain.Inventory
	err = db.Where("owner_id = ? AND item_name = ?", userID, itemName).First(&inventory).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, inventory.ItemAmount)
}

func TestSendCoinsE2E(t *testing.T) {
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

	senderID := uuid.New().String()
	receiverID := uuid.New().String() // Валидный UUID
	senderName := fmt.Sprintf("s_%d", time.Now().UnixNano())
	receiverUsername := fmt.Sprintf("r_%d", time.Now().UnixNano())
	createTestUser(t, db, senderID, senderName, 500)
	createTestUser(t, db, receiverID, receiverUsername, 100)

	token, err := jwtToken.Create(senderID, time.Now().Add(24*time.Hour).Unix())
	assert.NoError(t, err)

	authRepo := authRepository.NewAuthRepository(db)
	authUC := authUsecase.NewAuthUsecase(authRepo)
	authHandler := auth.NewAuthHandler(authUC, jwtToken)

	merchRepo := merchRepository.NewMerchRepository(db)
	merchUC := merchUsecase.NewMerchUsecase(merchRepo)
	merchHandler := merchController.NewMerchHandler(merchUC, jwtToken)

	router := mux.NewRouter()
	api := "/api"

	router.HandleFunc(api+"/auth", authHandler.LoginUser).Methods("POST")
	router.HandleFunc(api+"/info", merchHandler.GetUserMerchInformation).Methods("GET")
	router.HandleFunc(api+"/buy/{item}", merchHandler.BuyItem).Methods("GET")
	router.HandleFunc(api+"/sendCoin", merchHandler.SendCoins).Methods("POST")

	server := httptest.NewServer(router)
	defer server.Close()

	requestBody := domain.SentRequest{
		ToUser: receiverUsername,
		Amount: 200,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/sendCoin", bytes.NewBuffer(requestBodyBytes))
	assert.NoError(t, err)
	req.Header.Set("JWT-Token", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var sender domain.User
	err = db.Where("uuid = ?", senderID).First(&sender).Error
	assert.NoError(t, err)
	assert.Equal(t, 300, sender.Coins)

	var receiver domain.User
	err = db.Where("username = ?", receiverUsername).First(&receiver).Error
	assert.NoError(t, err)
	assert.Equal(t, 300, receiver.Coins)

	var transaction domain.Transaction
	err = db.Where("sender_id = ? AND receiver_id = ?", senderID, receiver.UUID).First(&transaction).Error
	assert.NoError(t, err)
	assert.Equal(t, 200, transaction.Amount)
}

func TestGetUserMerchInformationE2E(t *testing.T) {
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

	userID := uuid.New().String() // Валидный UUID
	username := fmt.Sprintf("u_%d", time.Now().UnixNano())
	createTestUser(t, db, userID, username, 500)

	createTestInventory(t, db, userID, "pen", 1)
	createTestInventory(t, db, userID, "hoody", 2)

	receiverID := uuid.New().String()
	receiverUsername := fmt.Sprintf("u_%d", time.Now().UnixNano())
	createTestUser(t, db, receiverID, receiverUsername, 100)
	createTestTransaction(t, db, userID, receiverID, 200)
	createTestTransaction(t, db, receiverID, userID, 50)

	token, err := jwtToken.Create(userID, time.Now().Add(24*time.Hour).Unix())
	assert.NoError(t, err)

	authRepo := authRepository.NewAuthRepository(db)
	authUC := authUsecase.NewAuthUsecase(authRepo)
	authHandler := auth.NewAuthHandler(authUC, jwtToken)

	merchRepo := merchRepository.NewMerchRepository(db)
	merchUC := merchUsecase.NewMerchUsecase(merchRepo)
	merchHandler := merchController.NewMerchHandler(merchUC, jwtToken)

	router := mux.NewRouter()
	api := "/api"

	router.HandleFunc(api+"/auth", authHandler.LoginUser).Methods("POST")
	router.HandleFunc(api+"/info", merchHandler.GetUserMerchInformation).Methods("GET")
	router.HandleFunc(api+"/buy/{item}", merchHandler.BuyItem).Methods("GET")
	router.HandleFunc(api+"/sendCoin", merchHandler.SendCoins).Methods("POST")

	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/info", nil)
	assert.NoError(t, err)
	req.Header.Set("JWT-Token", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response domain.UserInformationResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 500, response.Coins)

	assert.Len(t, response.Inventory, 2)
	for _, item := range response.Inventory {
		switch item.Type {
		case "pen":
			assert.Equal(t, 1, item.Quantity)
		case "hoody":
			assert.Equal(t, 2, item.Quantity)
		default:
			t.Errorf("Unexpected item type: %s", item.Type)
		}
	}

	assert.Len(t, response.CoinHistory.Sent, 1)
	assert.Len(t, response.CoinHistory.Received, 1)

	for _, sent := range response.CoinHistory.Sent {
		assert.Equal(t, receiverUsername, sent.ToUser)
		assert.Equal(t, 200, sent.Amount)
	}

	for _, received := range response.CoinHistory.Received {
		assert.Equal(t, receiverUsername, received.FromUser)
		assert.Equal(t, 50, received.Amount)
	}
}
