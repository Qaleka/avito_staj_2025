package repository

import (
	"avito_staj_2025/internal/service/logger"
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func TestAuthUser(t *testing.T) {
	logger.DBLogger = zap.NewNop()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	authRepo := NewAuthRepository(gormDB)
	ctx := context.Background()

	t.Run("Success - Existing User", func(t *testing.T) {
		username := "validUser"
		password := "password"
		hashedPassword := "hashedPassword"
		rows := sqlmock.NewRows([]string{"uuid", "username", "password"}).
			AddRow("user-123", username, hashedPassword)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(username, 1).
			WillReturnRows(rows)

		user, err := authRepo.AuthUser(ctx, username, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-123", user.UUID)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, hashedPassword, user.Password)
	})

	t.Run("Success - Create New User", func(t *testing.T) {
		username := "newUser"
		password := "hashedPassword"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."uuid" LIMIT $2`)).
			WithArgs(username, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("username","password","coins") VALUES ($1,$2,$3) RETURNING "uuid"`)).
			WithArgs("newUser", "hashedPassword", 1000).
			WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow("some-uuid"))
		mock.ExpectCommit()
		user, err := authRepo.AuthUser(ctx, username, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
		assert.Empty(t, user.Password)
	})

	t.Run("Fail - Error Creating User", func(t *testing.T) {
		username := "errorUser"
		password := "hashedPassword"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."uuid" LIMIT 1`)).
			WithArgs(username).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("username","password","coins") VALUES ($1,$2,$3) RETURNING "uuid"`)).
			WithArgs("newUser", "hashedPassword", 1000).
			WillReturnError(errors.New("failed to create user"))
		mock.ExpectRollback()
		user, err := authRepo.AuthUser(ctx, username, password)

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Fail - DB Error on Query", func(t *testing.T) {
		username := "brokenUser"
		password := "hashedPassword"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."uuid" LIMIT 1`)).
			WithArgs(username).
			WillReturnError(errors.New("database error"))

		user, err := authRepo.AuthUser(ctx, username, password)

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
