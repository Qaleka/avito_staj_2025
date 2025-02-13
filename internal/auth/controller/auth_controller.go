package controller

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/auth/usecase"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"encoding/json"
	"errors"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type AuthHandler struct {
	usecase  usecase.AuthUsecase
	jwtToken middleware.JwtTokenService
}

func NewAuthHandler(usecase usecase.AuthUsecase, jwtToken middleware.JwtTokenService) *AuthHandler {
	return &AuthHandler{
		usecase:  usecase,
		jwtToken: jwtToken,
	}
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := middleware.GetRequestID(r.Context())
	ctx, cancel := middleware.WithTimeout(r.Context())
	sanitizer := bluemonday.UGCPolicy()
	defer cancel()

	logger.AccessLogger.Info("Received LoginUser request",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
	)

	authHeader := r.Header.Get("JWT-Token")
	if authHeader != "" {
		logger.AccessLogger.Warn("jwt_token already exists",
			zap.String("request_id", requestID),
			zap.Error(errors.New("jwt_token already exists")),
		)
		h.handleError(w, errors.New("jwt_token already exists"), requestID)
		return
	}

	var creds domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		logger.AccessLogger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.handleError(w, err, requestID)
		return
	}

	creds.Username = sanitizer.Sanitize(creds.Username)
	creds.Password = sanitizer.Sanitize(creds.Password)

	userID, err := h.usecase.LoginUser(ctx, creds.Username, creds.Password)
	if err != nil {
		logger.AccessLogger.Error("Failed to login",
			zap.String("request_id", requestID),
			zap.Error(err))
		h.handleError(w, err, requestID)
		return
	}

	tokenExpTime := time.Now().Add(24 * time.Hour).Unix()
	jwtToken, err := h.jwtToken.Create(userID, tokenExpTime)
	if err != nil {
		logger.AccessLogger.Error("Failed to create JWT token",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.handleError(w, err, requestID)
		return
	}

	body := map[string]interface{}{
		"token": jwtToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.AccessLogger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duration := time.Since(start)
	logger.AccessLogger.Info("Completed LoginUser request",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK),
	)
}

func (h *AuthHandler) handleError(w http.ResponseWriter, err error, requestID string) {
	logger.AccessLogger.Error("Handling error",
		zap.String("request_id", requestID),
		zap.Error(err),
	)

	w.Header().Set("Content-Type", "application/json")
	errorResponse := map[string]string{"error": err.Error()}

	switch err.Error() {
	case "not correct username", "not correct password",
		"jwt_token already exists", "Input contains invalid characters",
		"Input exceeds character limit":
		w.WriteHeader(http.StatusBadRequest)
	case "invalid credentials":
		w.WriteHeader(http.StatusUnauthorized)
	case "failed to generate error response":
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if jsonErr := json.NewEncoder(w).Encode(errorResponse); jsonErr != nil {
		logger.AccessLogger.Error("Failed to encode error response",
			zap.String("request_id", requestID),
			zap.Error(jsonErr),
		)
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
	}
}
