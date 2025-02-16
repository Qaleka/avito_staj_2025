package controller

import (
	"avito_staj_2025/domain"
	"avito_staj_2025/internal/merch/usecase"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type MerchHandler struct {
	usecase  usecase.MerchUsecase
	jwtToken middleware.JwtTokenService
}

func NewMerchHandler(usecase usecase.MerchUsecase, jwtToken middleware.JwtTokenService) *MerchHandler {
	return &MerchHandler{
		usecase:  usecase,
		jwtToken: jwtToken,
	}
}

func (h *MerchHandler) SendCoins(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := middleware.GetRequestID(r.Context())
	ctx, cancel := middleware.WithTimeout(r.Context())
	sanitizer := bluemonday.UGCPolicy()
	defer cancel()

	logger.AccessLogger.Info("Received SendCoins request",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
	)

	authHeader := r.Header.Get("JWT-Token")
	if authHeader == "" {
		h.handleError(w, errors.New("Missing JWT-Token header"), requestID)
		return
	}

	tokenString := authHeader[len("Bearer "):]
	jwtToken, err := h.jwtToken.Validate(tokenString)
	if err != nil {
		h.handleError(w, errors.New("Invalid JWT token"), requestID)
		return
	}
	var data domain.SentRequest
	if err = json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.handleError(w, err, requestID)
		return
	}
	data.ToUser = sanitizer.Sanitize(data.ToUser)
	err = h.usecase.SendCoins(ctx, jwtToken.UserId, data.ToUser, data.Amount)
	if err != nil {
		h.handleError(w, err, requestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	duration := time.Since(start)
	logger.AccessLogger.Info("Completed SendCoins request",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK),
	)
}

func (h *MerchHandler) GetUserMerchInformation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := middleware.GetRequestID(r.Context())
	ctx, cancel := middleware.WithTimeout(r.Context())
	defer cancel()

	logger.AccessLogger.Info("Received GetUserMerchInformation request",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
	)

	authHeader := r.Header.Get("JWT-Token")
	if authHeader == "" {
		h.handleError(w, errors.New("Missing JWT-Token header"), requestID)
		return
	}

	tokenString := authHeader[len("Bearer "):]
	jwtToken, err := h.jwtToken.Validate(tokenString)
	if err != nil {
		h.handleError(w, errors.New("Invalid JWT token"), requestID)
		return
	}
	response, err := h.usecase.GetUserMerchInformation(ctx, jwtToken.UserId)
	if err != nil {
		h.handleError(w, err, requestID)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.handleError(w, err, requestID)
		return
	}
	duration := time.Since(start)
	logger.AccessLogger.Info("Completed GetUserMerchInformation request",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK))
}

func (h *MerchHandler) BuyItem(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := middleware.GetRequestID(r.Context())
	ctx, cancel := middleware.WithTimeout(r.Context())
	defer cancel()
	logger.AccessLogger.Info("Received BuyItem request",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
	)

	authHeader := r.Header.Get("JWT-Token")
	if authHeader == "" {
		h.handleError(w, errors.New("Missing JWT-Token header"), requestID)
		return
	}

	tokenString := authHeader[len("Bearer "):]
	jwtToken, err := h.jwtToken.Validate(tokenString)
	if err != nil {
		h.handleError(w, errors.New("Invalid JWT token"), requestID)
		return
	}
	itemName := mux.Vars(r)["item"]
	err = h.usecase.BuyItem(ctx, jwtToken.UserId, itemName)
	if err != nil {
		h.handleError(w, err, requestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	duration := time.Since(start)
	logger.AccessLogger.Info("Completed BuyItem request",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK))

}

func (h *MerchHandler) handleError(w http.ResponseWriter, err error, requestID string) {
	logger.AccessLogger.Error("Handling error",
		zap.String("request_id", requestID),
		zap.Error(err),
	)

	w.Header().Set("Content-Type", "application/json")
	errorResponse := map[string]string{"error": err.Error()}

	switch err.Error() {
	case "Input contains invalid characters", "Input exceeds character limit",
		"amount must be greater than 0", "item not found in merch types", "sender not found", "receiver not found",
		"not enough coins", "User not found":
		w.WriteHeader(http.StatusBadRequest)
	case "Invalid JWT token", "Missing JWT-Token header":
		w.WriteHeader(http.StatusUnauthorized)
	case "failed to start transaction", "failed to find sender", "failed to find receiver", "failed to update sender balance",
		"failed to update receiver balance", "failed to create transaction record", "failed to commit transaction",
		"failed to fetch user", "failed to fetch inventory", "failed to fetch sent transactions",
		"failed to fetch received transactions", "failed to update user balance", "failed to add new item to inventory",
		"failed to fetch inventory item", "failed to update inventory item":
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
