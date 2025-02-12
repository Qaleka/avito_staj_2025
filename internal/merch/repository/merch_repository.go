package repository

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type merchRepository struct {
	db *gorm.DB
}

func NewMerchRepository(db *gorm.DB) domain.MerchRepository {
	return &merchRepository{
		db: db,
	}
}

func (r *merchRepository) SendCoins(ctx context.Context, senderID string, receiverUsername string, amount int) error {
	requestID := middleware.GetRequestID(ctx)
	logger.DBLogger.Info("SendCoins called", zap.String("request_id", requestID), zap.String("receiverID", receiverUsername), zap.Int("amount", amount))

	tx := r.db.Begin()
	if tx.Error != nil {
		logger.DBLogger.Error("Failed to start transaction", zap.Error(tx.Error))
		return errors.New("failed to start transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sender domain.User
	if err := tx.Where("uuid = ?", senderID).First(&sender).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.DBLogger.Warn("User not found", zap.String("request_id", requestID), zap.String("sender_id", senderID))
			return errors.New("sender not found")
		}
		logger.DBLogger.Error("Failed to get user", zap.String("request_id", requestID), zap.String("sender_id", senderID))
		return errors.New("failed to find sender")
	}

	var receiver domain.User
	if err := tx.Where("username = ?", receiverUsername).First(&receiver).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.DBLogger.Warn("User not found", zap.String("request_id", requestID), zap.String("receiver_id", receiverUsername))
			return errors.New("receiver not found")
		}
		logger.DBLogger.Error("Failed to get user", zap.String("request_id", requestID), zap.String("receiver_id", receiverUsername))
		return errors.New("failed to find receiver")
	}

	if sender.Coins < amount {
		tx.Rollback()
		logger.DBLogger.Warn("Not enough coins", zap.String("request_id", requestID), zap.String("sender_id", senderID))
		return errors.New("not enough coins")
	}

	sender.Coins -= amount
	if err := tx.Save(&sender).Error; err != nil {
		tx.Rollback()
		logger.DBLogger.Error("Failed to update sender coins", zap.String("request_id", requestID), zap.String("sender_id", senderID))
		return errors.New("failed to update sender balance")
	}

	receiver.Coins += amount
	if err := tx.Save(&receiver).Error; err != nil {
		tx.Rollback()
		logger.DBLogger.Error("Failed to update receiver coins", zap.String("request_id", requestID), zap.String("receiver_id", receiver.UUID))
		return errors.New("failed to update receiver balance")
	}

	transaction := domain.Transaction{
		SenderID:   senderID,
		ReceiverID: receiver.UUID,
		Amount:     amount,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		logger.DBLogger.Error("Failed to create transaction", zap.String("request_id", requestID), zap.String("sender_id", senderID), zap.String("receiver_id", receiver.UUID))
		return errors.New("failed to create transaction record")
	}

	if err := tx.Commit().Error; err != nil {
		logger.DBLogger.Error("Failed to commit transaction", zap.String("request_id", requestID), zap.String("sender_id", senderID), zap.String("receiver_id", receiver.UUID))
		return errors.New("failed to commit transaction")
	}

	logger.DBLogger.Info("Successfully sent coins", zap.String("request_id", requestID), zap.String("sender_id", senderID), zap.String("receiver_id", receiver.UUID), zap.Int("amount", amount))
	return nil
}

func (r *merchRepository) GetUserMerchInformation(ctx context.Context, userID string) (domain.UserInformationResponse, error) {
	requestID := middleware.GetRequestID(ctx)
	logger.DBLogger.Info("GetUserMerchInformation called", zap.String("request_id", requestID), zap.String("user_id", userID))

	var user domain.User
	if err := r.db.Where("uuid = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.DBLogger.Warn("User not found", zap.String("request_id", requestID), zap.String("user_id", userID))
			return domain.UserInformationResponse{}, errors.New("user not found")
		}
		logger.DBLogger.Error("Failed to get user", zap.String("request_id", requestID), zap.String("user_id", userID))
		return domain.UserInformationResponse{}, errors.New("failed to fetch user")
	}

	var inventory []domain.Inventory
	if err := r.db.Where("owner_id = ?", userID).Find(&inventory).Error; err != nil {
		logger.DBLogger.Error("Failed to get inventory", zap.String("request_id", requestID), zap.String("user_id", userID))
		return domain.UserInformationResponse{}, errors.New("failed to fetch inventory")
	}

	var inventoryResponse []domain.InventoryResponse
	for _, item := range inventory {
		inventoryResponse = append(inventoryResponse, domain.InventoryResponse{
			Type:     item.ItemName,
			Quantity: item.ItemAmount,
		})
	}

	var transactions []domain.TransactionWithUsers
	if err := r.db.
		Table("transactions").
		Select("transactions.amount, sender.username AS sender_name, receiver.username AS receiver_name").
		Joins("JOIN users AS sender ON transactions.sender_id = sender.uuid").
		Joins("JOIN users AS receiver ON transactions.receiver_id = receiver.uuid").
		Where("transactions.sender_id = ? OR transactions.receiver_id = ?", userID, userID).
		Scan(&transactions).Error; err != nil {
		return domain.UserInformationResponse{}, errors.New("failed to fetch transactions")
	}

	var sentResponses []domain.SentResponse
	var receivedResponses []domain.ReceivedResponse
	for _, t := range transactions {
		if t.SenderName == user.Username {
			sentResponses = append(sentResponses, domain.SentResponse{
				ToUser: t.ReceiverName,
				Amount: t.Amount,
			})
		} else {
			receivedResponses = append(receivedResponses, domain.ReceivedResponse{
				FromUser: t.SenderName,
				Amount:   t.Amount,
			})
		}
	}

	response := domain.UserInformationResponse{
		Coins:     user.Coins,
		Inventory: inventoryResponse,
		CoinHistory: domain.CoinHistory{
			Sent:     sentResponses,
			Received: receivedResponses,
		},
	}

	logger.DBLogger.Info("Successfully get information", zap.String("request_id", requestID), zap.String("user_id", userID))
	return response, nil
}

func (r *merchRepository) BuyItem(ctx context.Context, userID string, itemName string, itemCost int) error {
	requestID := middleware.GetRequestID(ctx)
	logger.DBLogger.Info("BuyItem called", zap.String("request_id", requestID), zap.String("itemName", itemName), zap.String("user_id", userID))

	var user domain.User
	if err := r.db.Where("uuid = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.DBLogger.Warn("User not found", zap.String("request_id", requestID), zap.String("user_id", userID))
			return errors.New("User not found")
		}
		logger.DBLogger.Error("Failed to get user", zap.String("request_id", requestID), zap.String("user_id", userID))
		return errors.New("failed to fetch user")
	}

	if user.Coins < itemCost {
		logger.DBLogger.Warn("Not enough coins", zap.String("request_id", requestID), zap.String("user_id", userID))
		return errors.New("not enough coins")
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		logger.DBLogger.Error("Failed to start transaction", zap.String("request_id", requestID), zap.String("user_id", userID))
		return errors.New("failed to start transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user.Coins -= itemCost
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		logger.DBLogger.Error("Failed to update user coins", zap.String("request_id", requestID), zap.String("user_id", userID))
		return errors.New("failed to update user balance")
	}

	var inventoryItem domain.Inventory
	if err := tx.Where("owner_id = ? AND item_name = ?", userID, itemName).First(&inventoryItem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newItem := domain.Inventory{
				OwnerID:    userID,
				ItemName:   itemName,
				ItemAmount: 1,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				tx.Rollback()
				logger.DBLogger.Error("Failed to add new item to inventory", zap.String("request_id", requestID), zap.String("user_id", userID))
				return errors.New("failed to add new item to inventory")
			}
		} else {
			tx.Rollback()
			logger.DBLogger.Error("Failed to get inventory", zap.String("request_id", requestID), zap.String("user_id", userID))
			return errors.New("failed to fetch inventory item")
		}
	} else {
		inventoryItem.ItemAmount++
		if err := tx.Save(&inventoryItem).Error; err != nil {
			tx.Rollback()
			logger.DBLogger.Error("Failed to update inventory", zap.String("request_id", requestID), zap.String("user_id", userID))
			return errors.New("failed to update inventory item")
		}
	}

	transaction := domain.Transaction{
		SenderID:   userID,
		ReceiverID: "7d883678-7e5f-4a98-b77b-04989a62f3a6",
		Amount:     itemCost,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		logger.DBLogger.Error("Failed to create transaction", zap.String("request_id", requestID), zap.String("user_id", userID))
		return errors.New("failed to create transaction record")
	}

	if err := tx.Commit().Error; err != nil {
		logger.DBLogger.Error("Failed to commit", zap.String("request_id", requestID), zap.String("user_id", userID))
		return errors.New("failed to commit transaction")
	}

	logger.DBLogger.Info("Item successfully purchased", zap.String("request_id", requestID), zap.String("item_name", itemName), zap.String("user_id", userID))
	return nil
}
