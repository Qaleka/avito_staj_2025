package mocks

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/service/middleware"
	"context"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
)

// MockAuthRepository - мок репозитория аутентификации
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) AuthUser(ctx context.Context, username string, password string) (*domain.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockAuthUsecase - мок usecase для аутентификации
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) LoginUser(ctx context.Context, username string, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

// MockJwtTokenService - мок сервиса JWT
type MockJwtTokenService struct {
	mock.Mock
}

func (m *MockJwtTokenService) Create(userID string, tokenExpTime int64) (string, error) {
	args := m.Called(userID, tokenExpTime)
	return args.String(0), args.Error(1)
}

func (m *MockJwtTokenService) Validate(tokenString string) (*middleware.JwtCsrfClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) != nil {
		return args.Get(0).(*middleware.JwtCsrfClaims), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockJwtTokenService) ParseSecretGetter(token *jwt.Token) (interface{}, error) {
	args := m.Called(token)
	return args.Get(0), args.Error(1)
}
