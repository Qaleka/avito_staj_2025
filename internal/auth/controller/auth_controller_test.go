package controller

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/auth/mocks"
	"avito_staj_2025/internal/service/logger"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginUser(t *testing.T) {
	logger.AccessLogger = zap.NewNop()

	t.Run("Success - Valid Credentials", func(t *testing.T) {
		mockUsecase := new(mocks.MockAuthUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := AuthHandler{usecase: mockUsecase, jwtToken: mockJWT}

		credentials := domain.LoginRequest{Username: "validUser", Password: "validPassword"}
		requestBody, _ := json.Marshal(credentials)

		mockUsecase.On("LoginUser", mock.Anything, "validUser", "validPassword").Return("user-uuid", nil)
		mockJWT.On("Create", "user-uuid", mock.AnythingOfType("int64")).Return("validToken", nil)

		r, w := createTestRequest(http.MethodPost, "/auth", requestBody)
		h.LoginUser(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseBody map[string]string
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
		assert.Equal(t, "validToken", responseBody["token"])

		t.Cleanup(func() {
			mockUsecase.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	})

	t.Run("Failure - Invalid Credentials", func(t *testing.T) {
		mockUsecase := new(mocks.MockAuthUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := AuthHandler{usecase: mockUsecase, jwtToken: mockJWT}

		credentials := domain.LoginRequest{Username: "invalidUser", Password: "wrongPassword"}
		requestBody, _ := json.Marshal(credentials)

		mockUsecase.On("LoginUser", mock.Anything, "invalidUser", "wrongPassword").Return("", errors.New("invalid credentials"))

		r, w := createTestRequest(http.MethodPost, "/auth", requestBody)
		h.LoginUser(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		t.Cleanup(func() {
			mockUsecase.AssertExpectations(t)
		})
	})

	t.Run("Failure - JWT Creation Error", func(t *testing.T) {
		mockUsecase := new(mocks.MockAuthUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := AuthHandler{usecase: mockUsecase, jwtToken: mockJWT}

		credentials := domain.LoginRequest{Username: "validUser", Password: "validPassword"}
		requestBody, _ := json.Marshal(credentials)

		mockUsecase.On("LoginUser", mock.Anything, "validUser", "validPassword").Return("user-uuid", nil)
		mockJWT.On("Create", "user-uuid", mock.AnythingOfType("int64")).Return("", errors.New("jwt creation failed"))

		r, w := createTestRequest(http.MethodPost, "/auth", requestBody)
		h.LoginUser(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		t.Cleanup(func() {
			mockUsecase.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	})
}

func createTestRequest(method, url string, body []byte) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, bytes.NewReader(body))
	w := httptest.NewRecorder()
	return r, w
}
