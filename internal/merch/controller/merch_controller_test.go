package controller

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/merch/mocks"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendCoins(t *testing.T) {
	logger.AccessLogger = zap.NewNop()

	t.Run("Success - Coins Sent", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)

		requestBody := domain.SentRequest{ToUser: "receiver123", Amount: 100}
		body, _ := json.Marshal(requestBody)

		claims := &middleware.JwtCsrfClaims{UserId: "sender123", StandardClaims: jwt.StandardClaims{ExpiresAt: 86400}}
		mockJWT.On("Validate", "valid_token").Return(claims, nil)
		mockUsecase.On("SendCoins", mock.Anything, "sender123", "receiver123", 100).Return(nil)

		r, w := createTestRequest(http.MethodPost, "/api/sendCoin", body)
		r.Header.Set("JWT-Token", "Bearer valid_token")

		h.SendCoins(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		t.Cleanup(func() {
			mockUsecase.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	})

	t.Run("Failure - Missing JWT Token", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)

		r, w := createTestRequest(http.MethodPost, "/api/sendCoin", nil)
		h.SendCoins(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Failure - Invalid JWT Token", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)

		mockJWT.On("Validate", "invalid_token").Return(nil, errors.New("invalid token"))

		r, w := createTestRequest(http.MethodPost, "/api/sendCoin", nil)
		r.Header.Set("JWT-Token", "Bearer invalid_token")

		h.SendCoins(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestGetUserMerchInformation(t *testing.T) {
	logger.AccessLogger = zap.NewNop()

	t.Run("Success - Get Merch Info", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)

		claims := &middleware.JwtCsrfClaims{UserId: "user123", StandardClaims: jwt.StandardClaims{ExpiresAt: 86400}}
		mockJWT.On("Validate", "valid_token").Return(claims, nil)
		mockUsecase.On("GetUserMerchInformation", mock.Anything, "user123").Return(domain.UserInformationResponse{Coins: 500}, nil)

		r, w := createTestRequest(http.MethodGet, "/api/info", nil)
		r.Header.Set("JWT-Token", "Bearer valid_token")

		h.GetUserMerchInformation(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Failure - Missing JWT Token", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)

		r, w := createTestRequest(http.MethodGet, "/api/info", nil)
		h.GetUserMerchInformation(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestBuyItem(t *testing.T) {
	logger.AccessLogger = zap.NewNop()

	t.Run("Success - Item Purchased", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)
		item := "hoody"
		claims := &middleware.JwtCsrfClaims{UserId: "user123", StandardClaims: jwt.StandardClaims{ExpiresAt: 86400}}
		mockJWT.On("Validate", "valid_token").Return(claims, nil)
		mockUsecase.On("BuyItem", mock.Anything, "user123", item).Return(nil)

		r, w := createTestRequest(http.MethodGet, "/api/buy/"+item, nil)
		r.Header.Set("JWT-Token", "Bearer valid_token")
		r = mux.SetURLVars(r, map[string]string{"item": item})

		h.BuyItem(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Failure - Invalid JWT Token", func(t *testing.T) {
		mockUsecase := new(mocks.MockMerchUsecase)
		mockJWT := new(mocks.MockJwtTokenService)
		h := NewMerchHandler(mockUsecase, mockJWT)

		mockJWT.On("Validate", "invalid_token").Return(nil, errors.New("invalid token"))

		r, w := createTestRequest(http.MethodGet, "/api/buy/hoody", nil)
		r.Header.Set("JWT-Token", "Bearer invalid_token")

		h.BuyItem(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func createTestRequest(method, url string, body []byte) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, bytes.NewReader(body))
	w := httptest.NewRecorder()
	return r, w
}
