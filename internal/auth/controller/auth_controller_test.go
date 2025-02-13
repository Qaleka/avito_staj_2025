package controller

import (
	"avito_staj_2025/domain"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginUser(t *testing.T) {
	mockUsecase := new(mock.AuthUsecase)
	mockJWT := new(mock.JWTToken)

	h := controller.AuthHandler{
		usecase:  mockUsecase,
		jwtToken: mockJWT,
	}

	t.Run("Success - Valid Credentials", func(t *testing.T) {
		credentials := domain.LoginRequest{
			Username: "validUser",
			Password: "validPassword",
		}
		requestBody, _ := json.Marshal(credentials)

		mockUsecase.On("LoginUser", mock.Anything, "validUser", "validPassword").Return("user-uuid", nil)
		mockJWT.On("Create", "user-uuid", mock.AnythingOfType("int64")).Return("validToken", nil)

		r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
		w := httptest.NewRecorder()

		h.LoginUser(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseBody map[string]string
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
		assert.Equal(t, "validToken", responseBody["token"])
		mockUsecase.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("Failure - Invalid Credentials", func(t *testing.T) {
		credentials := domain.LoginRequest{
			Username: "invalidUser",
			Password: "wrongPassword",
		}
		requestBody, _ := json.Marshal(credentials)

		mockUsecase.On("LoginUser", mock.Anything, "invalidUser", "wrongPassword").Return("", errors.New("invalid credentials"))

		r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
		w := httptest.NewRecorder()

		h.LoginUser(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("Failure - JWT Creation Error", func(t *testing.T) {
		credentials := domain.LoginRequest{
			Username: "validUser",
			Password: "validPassword",
		}
		requestBody, _ := json.Marshal(credentials)

		mockUsecase.On("LoginUser", mock.Anything, "validUser", "validPassword").Return("user-uuid", nil)
		mockJWT.On("Create", "user-uuid", mock.AnythingOfType("int64")).Return("", errors.New("jwt creation failed"))

		r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
		w := httptest.NewRecorder()

		h.LoginUser(w, r)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})
}
