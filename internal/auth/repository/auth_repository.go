package repository

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) domain.AuthRepository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) AuthUser(ctx context.Context, username string, password string) (*domain.User, error) {
	requestID := middleware.GetRequestID(ctx)
	var user domain.User
	logger.DBLogger.Info("AuthUser called", zap.String("request_id", requestID), zap.String("username", username))
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			user.Username = username
			user.Password = password
			user.Coins = 1000
			err = r.db.Create(&user).Error
			if err != nil {
				logger.DBLogger.Error("Error creating user", zap.String("username", username), zap.Error(err))
				return nil, err
			}
			logger.DBLogger.Info("Successfully create user", zap.String("request_id", requestID), zap.String("username", username))
			return &domain.User{UUID: user.UUID, Username: username, Password: ""}, nil
		}
		logger.DBLogger.Error("Error getting user", zap.String("request_id", requestID), zap.String("username", username), zap.Error(err))
		return nil, err
	}
	logger.DBLogger.Info("Successfully auth user", zap.String("request_id", requestID), zap.String("username", username))
	return &domain.User{UUID: user.UUID, Username: user.Username, Password: user.Password}, nil
}
