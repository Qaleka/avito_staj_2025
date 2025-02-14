package repository

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/service/logger"
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func TestSendCoins(t *testing.T) {
	logger.DBLogger = zap.NewNop()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewMerchRepository(gormDB)
	ctx := context.Background()

	senderID := "sender-uuid"
	receiverUsername := "receiverUser"
	amount := 100

	t.Run("Success - Send Coins", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE uuid = \$1 ORDER BY "users"."uuid" LIMIT \$2`).
			WithArgs(senderID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"uuid", "coins"}).AddRow(senderID, 200))

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"."uuid" LIMIT \$2`).
			WithArgs(receiverUsername, 1).
			WillReturnRows(sqlmock.NewRows([]string{"uuid", "username", "coins"}).AddRow("receiver-uuid", receiverUsername, 50))

		mock.ExpectExec(`UPDATE \"users\" SET \"coins\"=\$1 WHERE uuid = \$2`).
			WithArgs(100, senderID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`UPDATE \"users\" SET \"coins\"=\$1 WHERE uuid = \$2`).
			WithArgs(150, "receiver-uuid").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(`INSERT INTO \"transactions\"`).
			WithArgs(senderID, "receiver-uuid", amount).
			WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow("transaction-uuid"))

		mock.ExpectCommit()

		err := repo.SendCoins(ctx, senderID, receiverUsername, amount)
		assert.NoError(t, err)
	})

	t.Run("Fail - Sender Not Found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE uuid = \$1 ORDER BY "users"."uuid" LIMIT \$2`).
			WithArgs(senderID, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		err := repo.SendCoins(ctx, senderID, receiverUsername, amount)
		assert.Error(t, err)
		assert.Equal(t, "sender not found", err.Error())
	})

	t.Run("Fail - Not Enough Coins", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE uuid = \$1 ORDER BY "users"."uuid" LIMIT \$2`).
			WithArgs(senderID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"uuid", "coins"}).AddRow(senderID, 50))

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"."uuid" LIMIT \$2`).
			WithArgs(receiverUsername, 1).
			WillReturnRows(sqlmock.NewRows([]string{"uuid", "coins"}).AddRow(receiverUsername, 100))
		mock.ExpectRollback()

		err := repo.SendCoins(ctx, senderID, receiverUsername, amount)
		assert.Error(t, err)
		assert.Equal(t, "not enough coins", err.Error())
	})
}

func TestGetUserMerchInformation(t *testing.T) {
	logger.DBLogger = zap.NewNop()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewMerchRepository(gormDB)
	ctx := context.Background()
	userID := "user-uuid"

	t.Run("Success - Fetch User Merch Information", func(t *testing.T) {
		userRows := sqlmock.NewRows([]string{"uuid", "coins", "username"}).
			AddRow(userID, 500, "Alice")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnRows(userRows)

		inventoryRows := sqlmock.NewRows([]string{"owner_id", "item_name", "item_amount"}).
			AddRow(userID, "sword", 2).
			AddRow(userID, "shield", 1)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inventories" WHERE owner_id = $1`)).
			WithArgs(userID).
			WillReturnRows(inventoryRows)

		transactionsRows := sqlmock.NewRows([]string{"amount", "sender_name", "receiver_name"}).
			AddRow(100, "Alice", "Bob").
			AddRow(50, "Bob", "Alice")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT transactions.amount, sender.username AS sender_name, receiver.username AS receiver_name FROM "transactions" JOIN users AS sender ON transactions.sender_id = sender.uuid JOIN users AS receiver ON transactions.receiver_id = receiver.uuid WHERE transactions.sender_id = $1 OR transactions.receiver_id = $2`)).
			WithArgs(userID, userID).
			WillReturnRows(transactionsRows)

		response, err := repo.GetUserMerchInformation(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 500, response.Coins)
		assert.Len(t, response.Inventory, 2)
		assert.Len(t, response.CoinHistory.Sent, 1)
		assert.Len(t, response.CoinHistory.Received, 1)
	})

	t.Run("Fail - User Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		response, err := repo.GetUserMerchInformation(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		assert.Equal(t, domain.UserInformationResponse{}, response)
	})
}

func TestBuyItem(t *testing.T) {
	logger.DBLogger = zap.NewNop()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewMerchRepository(gormDB)
	ctx := context.Background()
	userID := "user-uuid"
	itemName := "sword"
	itemCost := 100

	t.Run("Success - Buy New Item", func(t *testing.T) {
		userRows := sqlmock.NewRows([]string{"uuid", "coins"}).
			AddRow(userID, 500)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnRows(userRows)

		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "coins"=$1 WHERE uuid = $2`)).
			WithArgs(400, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inventories" WHERE owner_id = $1 AND item_name = $2 ORDER BY "inventories"."id" LIMIT $3`)).
			WithArgs(userID, itemName, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Исправлено: ExpectQuery вместо ExpectExec
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inventories" ("owner_id","item_name","item_amount") VALUES ($1,$2,$3)`)).
			WithArgs(userID, itemName, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectCommit()

		err := repo.BuyItem(ctx, userID, itemName, itemCost)

		assert.NoError(t, err)
	})

	t.Run("Fail - Not Enough Coins", func(t *testing.T) {
		userRows := sqlmock.NewRows([]string{"uuid", "coins"}).
			AddRow(userID, 50)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnRows(userRows)

		err := repo.BuyItem(ctx, userID, itemName, itemCost)

		assert.Error(t, err)
		assert.Equal(t, "not enough coins", err.Error())
	})

	t.Run("Fail - User Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		err := repo.BuyItem(ctx, userID, itemName, itemCost)

		assert.Error(t, err)
		assert.Equal(t, "User not found", err.Error())
	})
}
