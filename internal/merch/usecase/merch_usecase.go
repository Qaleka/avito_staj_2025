package usecase

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"context"
	"errors"
	"go.uber.org/zap"
	"regexp"
)

type MerchUsecase interface {
	SendCoins(ctx context.Context, senderID string, receiverUsername string, amount int) error
	GetUserMerchInformation(ctx context.Context, userID string) (domain.UserInformationResponse, error)
	BuyItem(ctx context.Context, userID string, itemName string) error
}

type merchUsecase struct {
	merchRepository domain.MerchRepository
}

func NewMerchUsecase(merchRepository domain.MerchRepository) MerchUsecase {
	return &merchUsecase{
		merchRepository: merchRepository,
	}
}

func (uc *merchUsecase) SendCoins(ctx context.Context, senderID string, receiverUsername string, amount int) error {
	const maxLen = 255
	requestID := middleware.GetRequestID(ctx)
	validCharPattern := regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-_]*$`)
	if !validCharPattern.MatchString(senderID) {
		logger.AccessLogger.Warn("Input contains invalid characters", zap.String("request_id", requestID))
		return errors.New("Input contains invalid characters")
	}

	if len(senderID) > maxLen {
		logger.AccessLogger.Warn("Input exceeds character limit", zap.String("request_id", requestID))
		return errors.New("Input exceeds character limit")
	}

	if amount <= 0 {
		logger.AccessLogger.Warn("coins needs to be positive", zap.String("request_id", requestID))
		return errors.New("amount must be greater than 0")
	}

	err := uc.merchRepository.SendCoins(ctx, senderID, receiverUsername, amount)
	if err != nil {
		return err
	}

	return nil
}

func (uc *merchUsecase) GetUserMerchInformation(ctx context.Context, userID string) (domain.UserInformationResponse, error) {
	const maxLen = 255
	requestID := middleware.GetRequestID(ctx)
	validCharPattern := regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-_]*$`)
	if !validCharPattern.MatchString(userID) {
		logger.AccessLogger.Warn("Input contains invalid characters", zap.String("request_id", requestID))
		return domain.UserInformationResponse{}, errors.New("Input contains invalid characters")
	}

	if len(userID) > maxLen {
		logger.AccessLogger.Warn("Input exceeds character limit", zap.String("request_id", requestID))
		return domain.UserInformationResponse{}, errors.New("Input exceeds character limit")
	}

	response, err := uc.merchRepository.GetUserMerchInformation(ctx, userID)
	if err != nil {
		return domain.UserInformationResponse{}, err
	}
	return response, nil
}

func (uc *merchUsecase) BuyItem(ctx context.Context, userID string, itemName string) error {
	const maxLen = 255
	requestID := middleware.GetRequestID(ctx)
	validCharPattern := regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-_]*$`)
	if !validCharPattern.MatchString(userID) {
		logger.AccessLogger.Warn("Input contains invalid characters", zap.String("request_id", requestID))
		return errors.New("Input contains invalid characters")
	}

	if len(userID) > maxLen {
		logger.AccessLogger.Warn("Input exceeds character limit", zap.String("request_id", requestID))
		return errors.New("Input exceeds character limit")
	}

	itemCost, exists := domain.MerchTypes[itemName]
	if !exists {
		logger.AccessLogger.Warn("Item not found", zap.String("request_id", requestID), zap.String("itemName", itemName))
		return errors.New("item not found in merch types")
	}

	err := uc.merchRepository.BuyItem(ctx, userID, itemName, itemCost)
	if err != nil {
		return err
	}
	return nil
}
