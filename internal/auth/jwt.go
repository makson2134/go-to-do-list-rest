package auth

import (
	"fmt"
	"time"
	"to-do-list/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

//Идея реализация взята из статьи https://ru.hexlet.io/courses/go-web-development/lessons/auth/theory_unit

type TokenManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewTokenManager(secretKey string, tokenDuration time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (tm *TokenManager) GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(tm.tokenDuration).Unix(),
		"iat": time.Now().Unix(),
		"uid": user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(tm.secretKey))
	if err != nil {
		return "", fmt.Errorf("auth: failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (tm *TokenManager) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.secretKey), nil
	})

	if err != nil {
		return 0, fmt.Errorf("auth: failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("auth: invalid token")
	}

	// В JWT числа декодируются как float64
	userIDFloat, ok := claims["uid"].(float64)
	if !ok {
		return 0, fmt.Errorf("auth: invalid user id type in token")
	}
	return uint(userIDFloat), nil
}
