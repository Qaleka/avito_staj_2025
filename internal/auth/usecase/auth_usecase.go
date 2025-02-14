package usecase

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"avito_staj_2025/internal/service/validation"
	"context"
	"errors"
	"go.uber.org/zap"
)

type AuthUsecase interface {
	LoginUser(ctx context.Context, username string, password string) (string, error)
}

type authUsecase struct {
	authRepository domain.AuthRepository
}

func NewAuthUsecase(authRepository domain.AuthRepository) AuthUsecase {
	return &authUsecase{
		authRepository: authRepository,
	}
}

func (uc *authUsecase) LoginUser(ctx context.Context, username string, password string) (string, error) {
	requestID := middleware.GetRequestID(ctx)
	const maxLen = 100
	if len(username) > maxLen || len(password) > maxLen {
		logger.AccessLogger.Warn("Input exceeds character limit", zap.String("request_id", requestID))
		return "", errors.New("Input exceeds character limit")
	}
	if !validation.ValidateLogin(username) {
		logger.AccessLogger.Warn("not correct username", zap.String("request_id", requestID))
		return "", errors.New("not correct username")
	}
	if !validation.ValidatePassword(password) {
		logger.AccessLogger.Warn("not corrects password", zap.String("request_id", requestID))
		return "", errors.New("not correct password")
	}

	user, err := uc.authRepository.AuthUser(ctx, username, password)
	if err != nil {
		return "", err
	}

	if user.Password != "" && user.Password != password {
		return "", errors.New("invalid credentials")
	}

	return user.UUID, nil
}
