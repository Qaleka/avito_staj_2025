package mocks

import (
	"avito_staj_2025/domain"
	"context"
	"github.com/stretchr/testify/mock"
)

type MockMerchUsecase struct {
	mock.Mock
}

func (m *MockMerchUsecase) SendCoins(ctx context.Context, senderID string, receiverUsername string, amount int) error {
	args := m.Called(ctx, senderID, receiverUsername, amount)
	return args.Error(0)
}

func (m *MockMerchUsecase) GetUserMerchInformation(ctx context.Context, userID string) (domain.UserInformationResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.UserInformationResponse), args.Error(1)
}

func (m *MockMerchUsecase) BuyItem(ctx context.Context, userID string, itemName string) error {
	args := m.Called(ctx, userID, itemName)
	return args.Error(0)
}

// Mock для MerchRepository
type MockMerchRepository struct {
	mock.Mock
}

func (m *MockMerchRepository) SendCoins(ctx context.Context, senderID string, receiverUsername string, amount int) error {
	args := m.Called(ctx, senderID, receiverUsername, amount)
	return args.Error(0)
}

func (m *MockMerchRepository) GetUserMerchInformation(ctx context.Context, userID string) (domain.UserInformationResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.UserInformationResponse), args.Error(1)
}

func (m *MockMerchRepository) BuyItem(ctx context.Context, userID string, itemName string, itemCost int) error {
	args := m.Called(ctx, userID, itemName, itemCost)
	return args.Error(0)
}

// Mock для JWT
type MockJWT struct {
	mock.Mock
}

func (m *MockJWT) GenerateToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWT) ValidateToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}
