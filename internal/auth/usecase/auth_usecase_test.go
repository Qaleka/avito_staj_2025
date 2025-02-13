package usecase

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/auth/mocks"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestLoginUser(t *testing.T) {
	logger.AccessLogger = zap.NewNop()

	mockRepo := new(mocks.MockAuthRepository)
	authUC := NewAuthUsecase(mockRepo)

	ctx := context.Background()
	validUsername := "validUser"
	validPassword := "Secure123!"
	tooLongString := strings.Repeat("a", 101)
	invalidUsername := "/~~~~~~~"
	invalidPassword := "short"

	hashedPassword, _ := middleware.HashPassword(validPassword)

	// Успешная аутентификация
	mockRepo.On("AuthUser", mock.Anything, validUsername, mock.MatchedBy(func(pwd string) bool {
		return middleware.CheckPassword(pwd, validPassword)
	})).Return(&domain.User{UUID: "user-123", Password: ""}, nil)

	t.Run("Success", func(t *testing.T) {
		userID, err := authUC.LoginUser(ctx, validUsername, validPassword)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", userID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockRepo.On("AuthUser", mock.Anything, validUsername, mock.Anything).
			Return(&domain.User{UUID: "user-123", Password: hashedPassword}, nil)

		userID, err := authUC.LoginUser(ctx, validUsername, "WrongPass123!")

		assert.Error(t, err)
		assert.Empty(t, userID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Input Exceeds Character Limit", func(t *testing.T) {
		userID, err := authUC.LoginUser(ctx, tooLongString, validPassword)
		assert.Error(t, err)
		assert.Equal(t, "Input exceeds character limit", err.Error())
		assert.Empty(t, userID)
	})

	t.Run("Invalid Username Format", func(t *testing.T) {
		userID, err := authUC.LoginUser(ctx, invalidUsername, validPassword)
		assert.Error(t, err)
		assert.Equal(t, "not correct username", err.Error())
		assert.Empty(t, userID)
	})

	t.Run("Invalid Password Format", func(t *testing.T) {
		userID, err := authUC.LoginUser(ctx, validUsername, invalidPassword)
		assert.Error(t, err)
		assert.Equal(t, "not correct password", err.Error())
		assert.Empty(t, userID)
	})
}
