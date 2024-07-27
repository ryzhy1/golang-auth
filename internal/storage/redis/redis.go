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
	redisClient := redis.NewClient(&redis.Options{
		Addr:     connStr,       // Адрес и порт Redis сервера
		Password: redisPassword, // Пароль, если он установлен
		DB:       redisDbNumber, // Номер базы данных Redis
	})
	return &Storage{db: redisClient}, nil
}

func (s *Storage) SaveUserCache(ctx context.Context, login, token string, duration time.Duration) error {
	const op = "storage.Redis.SaveToken"

	err := s.db.Set(ctx, login, token, duration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return err
}

func (s *Storage) InvalidateToken(ctx context.Context, token string) error {
	const op = "storage.Redis.InvalidateToken"

	err := s.db.Del(ctx, token).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return err
}

func (s *Storage) IsTokenValid(ctx context.Context, token string) bool {
	const op = "storage.Redis.IsTokenValid"

	result, err := s.db.Exists(ctx, token).Result()
	if err != nil {
		_ = fmt.Errorf("%s: %w", op, err)
		return false
	}
	return result == 1
}

func (s *Storage) LogoutUser(ctx context.Context, login, token string) (bool, error) {
	const op = "storage.Redis.LogoutUser"

	user, err := s.db.Exists(ctx, login).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if user == 0 {
		return false, fmt.Errorf("%s: %w", op, storage.ErrNoActiveSession)
	}

	if err = s.db.Del(ctx, login, token).Err(); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}
