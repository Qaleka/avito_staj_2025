package usecase

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/merch/mocks"
	"avito_staj_2025/internal/service/logger"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestSendCoins(t *testing.T) {
	logger.AccessLogger = zap.NewNop()
	mockRepo := new(mocks.MockMerchRepository)
	uc := NewMerchUsecase(mockRepo)

	ctx := context.Background()
	validSender := "user123"
	validReceiver := "receiver456"
	validAmount := 100
	invalidSender := "/invalid~user"
	tooLongSender := strings.Repeat("a", 256)
	negativeAmount := -50

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("SendCoins", ctx, validSender, validReceiver, validAmount).Return(nil)

		err := uc.SendCoins(ctx, validSender, validReceiver, validAmount)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Sender ID", func(t *testing.T) {
		err := uc.SendCoins(ctx, invalidSender, validReceiver, validAmount)
		assert.Error(t, err)
		assert.Equal(t, "Input contains invalid characters", err.Error())
	})

	t.Run("Sender ID Too Long", func(t *testing.T) {
		err := uc.SendCoins(ctx, tooLongSender, validReceiver, validAmount)
		assert.Error(t, err)
		assert.Equal(t, "Input exceeds character limit", err.Error())
	})

	t.Run("Negative Amount", func(t *testing.T) {
		err := uc.SendCoins(ctx, validSender, validReceiver, negativeAmount)
		assert.Error(t, err)
		assert.Equal(t, "amount must be greater than 0", err.Error())
	})
}

func TestGetUserMerchInformation(t *testing.T) {
	logger.AccessLogger = zap.NewNop()
	mockRepo := new(mocks.MockMerchRepository)
	uc := NewMerchUsecase(mockRepo)

	ctx := context.Background()
	validUserID := "user123"
	invalidUserID := "/invalid~user"
	tooLongUserID := strings.Repeat("a", 256)
	expectedResponse := domain.UserInformationResponse{
		Coins: 200,
		Inventory: []domain.InventoryResponse{
			{Type: "sword", Quantity: 1},
			{Type: "shield", Quantity: 2},
		},
		CoinHistory: domain.CoinHistory{
			Received: []domain.ReceivedResponse{
				{FromUser: "friend1", Amount: 50},
				{FromUser: "friend2", Amount: 100},
			},
			Sent: []domain.SentResponse{
				{ToUser: "player1", Amount: 30},
				{ToUser: "player2", Amount: 20},
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetUserMerchInformation", ctx, validUserID).Return(expectedResponse, nil)

		response, err := uc.GetUserMerchInformation(ctx, validUserID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid User ID", func(t *testing.T) {
		_, err := uc.GetUserMerchInformation(ctx, invalidUserID)
		assert.Error(t, err)
		assert.Equal(t, "Input contains invalid characters", err.Error())
	})

	t.Run("User ID Too Long", func(t *testing.T) {
		_, err := uc.GetUserMerchInformation(ctx, tooLongUserID)
		assert.Error(t, err)
		assert.Equal(t, "Input exceeds character limit", err.Error())
	})
}

func TestBuyItem(t *testing.T) {
	logger.AccessLogger = zap.NewNop()
	mockRepo := new(mocks.MockMerchRepository)
	uc := NewMerchUsecase(mockRepo)

	ctx := context.Background()
	validUserID := "user123"
	validItem := "sword"
	invalidUserID := "/invalid~user"
	tooLongUserID := strings.Repeat("a", 256)
	nonExistentItem := "unknown_item"

	// Определяем MerchTypes (может быть в другом месте)
	domain.MerchTypes = map[string]int{
		"sword":  100,
		"shield": 150,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("BuyItem", ctx, validUserID, validItem, 100).Return(nil)

		err := uc.BuyItem(ctx, validUserID, validItem)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid User ID", func(t *testing.T) {
		err := uc.BuyItem(ctx, invalidUserID, validItem)
		assert.Error(t, err)
		assert.Equal(t, "Input contains invalid characters", err.Error())
	})

	t.Run("User ID Too Long", func(t *testing.T) {
		err := uc.BuyItem(ctx, tooLongUserID, validItem)
		assert.Error(t, err)
		assert.Equal(t, "Input exceeds character limit", err.Error())
	})

	t.Run("Item Not Found", func(t *testing.T) {
		err := uc.BuyItem(ctx, validUserID, nonExistentItem)
		assert.Error(t, err)
		assert.Equal(t, "item not found in merch types", err.Error())
	})
}
