package redis

import (
	"AuthService/internal/storage"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type Storage struct {
	db *redis.Client
}

// InitRedis инициализирует клиент Redis.
func InitRedis(connStr, redisPassword string, redisDbNumber int) (*Storage, error) {
	const op = "storage.redis.InitRedis"

	redisClient := redis.NewClient(&redis.Options{
		Addr:     connStr,
		Password: redisPassword,
		DB:       redisDbNumber,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: redisClient}, nil
}

func (s *Storage) SaveUserCache(ctx context.Context, userID, token string, duration time.Duration) error {
	const op = "storage.Redis.SaveToken"

	err := s.db.Set(ctx, token, userID, duration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return err
}

func (s *Storage) LogoutUser(ctx context.Context, token string) (bool, error) {
	const op = "storage.Redis.LogoutUser"

	user, err := s.db.Exists(ctx, token).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if user == 0 {
		return false, fmt.Errorf("%s: %w", op, storage.ErrNoActiveSession)
	}

	if err = s.db.Del(ctx, token).Err(); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	const op = "storage.Redis.DeleteRefreshToken"

	err := s.db.Del(ctx, refreshToken).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func (s *Storage) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (userID string, err error) {
	const op = "storage.Redis.GetUserIDByRefreshToken"

	userID, err = s.db.Get(ctx, refreshToken).Result()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}
