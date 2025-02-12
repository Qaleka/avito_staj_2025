package middleware

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt"
)

type JwtTokenService interface {
	Create(userID string, tokenExpTime int64) (string, error)
	Validate(tokenString string) (*JwtCsrfClaims, error)
	ParseSecretGetter(token *jwt.Token) (interface{}, error)
}

type JwtToken struct {
	Secret []byte
}

func NewJwtToken(secret string) (JwtTokenService, error) {
	return &JwtToken{
		Secret: []byte(secret),
	}, nil
}

type JwtCsrfClaims struct {
	UserId string `json:"userID"`
	jwt.StandardClaims
}

func (tk *JwtToken) Create(userID string, tokenExpTime int64) (string, error) {
	data := JwtCsrfClaims{
		UserId: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpTime,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(tk.Secret)
}

func (tk *JwtToken) Validate(tokenString string) (*JwtCsrfClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtCsrfClaims{}, tk.ParseSecretGetter)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JwtCsrfClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Проверка срока действия (дополнительно)
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

func (tk *JwtToken) ParseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	return tk.Secret, nil
}
