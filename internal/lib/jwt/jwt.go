package jwt

import (
	"AuthService/internal/domain/models"
	"AuthService/internal/storage/redis"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

func NewToken(user *models.User, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(tokenTTL).Unix()

	if os.Getenv("JWT_SECRET") == "" {
		return "", fmt.Errorf("jwt secret is empty")
	}

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func RefreshToken(ctx context.Context, redisStorage *redis.Storage, refreshTokenString string) (string, error) {
	token, err := VerifyToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	userID, ok := claims["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user_id in token claims")
	}

	if !redisStorage.IsTokenValid(ctx, refreshTokenString) {
		return "", fmt.Errorf("refresh token is not valid")
	}

	err = redisStorage.InvalidateToken(ctx, refreshTokenString)
	if err != nil {
		return "", err
	}

	return NewToken(&models.User{ID: userID}, 15*time.Minute)
}
