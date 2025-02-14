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
	if err := tx.Model(&domain.User{}).Where("uuid = ?", senderID).Update("coins", sender.Coins).Error; err != nil {
		tx.Rollback()
		logger.DBLogger.Error("Failed to update sender coins", zap.String("request_id", requestID), zap.String("sender_id", senderID))
		return errors.New("failed to update sender balance")
	}

	receiver.Coins += amount
	if err := tx.Model(&domain.User{}).Where("uuid = ?", receiver.UUID).Update("coins", receiver.Coins).Error; err != nil {
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

	var response domain.UserInformationResponse

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		var user domain.User
		if err := tx.Where("uuid = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.DBLogger.Warn("User not found", zap.String("request_id", requestID), zap.String("user_id", userID))
				return errors.New("user not found")
			}
			logger.DBLogger.Error("Failed to get user", zap.String("request_id", requestID), zap.Error(err))
			return errors.New("failed to fetch user")
		}

		var inventory []domain.Inventory
		if err := tx.Where("owner_id = ?", userID).Find(&inventory).Error; err != nil {
			logger.DBLogger.Error("Failed to get inventory", zap.String("request_id", requestID), zap.Error(err))
			return errors.New("failed to fetch inventory")
		}

		var transactions []domain.TransactionWithUsers
		if err := tx.
			Table("transactions").
			Select("transactions.amount, sender.username AS sender_name, receiver.username AS receiver_name").
			Joins("JOIN users AS sender ON transactions.sender_id = sender.uuid").
			Joins("JOIN users AS receiver ON transactions.receiver_id = receiver.uuid").
			Where("transactions.sender_id = ? OR transactions.receiver_id = ?", userID, userID).
			Scan(&transactions).Error; err != nil {
			return errors.New("failed to fetch transactions")
		}

		response = domain.UserInformationResponse{
			Coins:     user.Coins,
			Inventory: make([]domain.InventoryResponse, len(inventory)),
			CoinHistory: domain.CoinHistory{
				Sent:     make([]domain.SentResponse, 0),
				Received: make([]domain.ReceivedResponse, 0),
			},
		}

		for i, item := range inventory {
			response.Inventory[i] = domain.InventoryResponse{
				Type:     item.ItemName,
				Quantity: item.ItemAmount,
			}
		}

		for _, t := range transactions {
			if t.SenderName == user.Username {
				response.CoinHistory.Sent = append(response.CoinHistory.Sent, domain.SentResponse{
					ToUser: t.ReceiverName,
					Amount: t.Amount,
				})
			} else {
				response.CoinHistory.Received = append(response.CoinHistory.Received, domain.ReceivedResponse{
					FromUser: t.SenderName,
					Amount:   t.Amount,
				})
			}
		}

		return nil
	}); err != nil {
		return domain.UserInformationResponse{}, err
	}

	logger.DBLogger.Info("Successfully get information", zap.String("request_id", requestID), zap.String("user_id", userID))
	return response, nil
}

func (r *merchRepository) BuyItem(ctx context.Context, userID string, itemName string, itemCost int) error {
	requestID := middleware.GetRequestID(ctx)
	logger.DBLogger.Info("BuyItem called", zap.String("request_id", requestID), zap.String("itemName", itemName), zap.String("user_id", userID))

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		var user domain.User
		if err := tx.Where("uuid = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.DBLogger.Warn("User not found", zap.String("request_id", requestID), zap.String("user_id", userID))
				return errors.New("user not found")
			}
			logger.DBLogger.Error("Failed to get user", zap.String("request_id", requestID), zap.Error(err))
			return errors.New("failed to fetch user")
		}

		if user.Coins < itemCost {
			logger.DBLogger.Warn("Not enough coins", zap.String("request_id", requestID), zap.String("user_id", userID))
			return errors.New("not enough coins")
		}

		if err := tx.Model(&domain.User{}).Where("uuid = ?", userID).Update("coins", user.Coins-itemCost).Error; err != nil {
			logger.DBLogger.Error("Failed to update user coins", zap.String("request_id", requestID), zap.Error(err))
			return errors.New("failed to update user balance")
		}

		if err := tx.Exec(`
			INSERT INTO inventories (owner_id, item_name, item_amount)
			VALUES (?, ?, 1)
			ON CONFLICT (owner_id, item_name)
			DO UPDATE SET item_amount = inventories.item_amount + 1
		`, userID, itemName).Error; err != nil {
			logger.DBLogger.Error("Failed to update inventory", zap.String("request_id", requestID), zap.Error(err))
			return errors.New("failed to update inventory")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.DBLogger.Info("Item successfully purchased", zap.String("request_id", requestID), zap.String("item_name", itemName), zap.String("user_id", userID))
	return nil
}
