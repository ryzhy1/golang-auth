package auth

import (
	"AuthService/internal/domain/models"
	"AuthService/internal/lib/jwt"
	"AuthService/internal/storage"
	"AuthService/middlewares"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, id uuid.UUID, login, email string, password []byte) (uid string, err error)
	CheckUsernameIsAvailable(ctx context.Context, login string) (status bool, err error)
	CheckEmailIsAvailable(ctx context.Context, email string) (status bool, err error)
}

type UserProvider interface {
	GetUser(ctx context.Context, inputType, input string) (user *models.User, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// New return a new instance of the Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Register(ctx context.Context, login, email, password string) (userID string, err error) {
	const op = "auth.Register"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	if status := middlewares.CheckRegister(login, email, password); status != true {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if status, err := a.userSaver.CheckUsernameIsAvailable(ctx, login); status != true || err != nil {
		if err != nil {
			log.Error("this username already taken", err)
		}

		return "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
	}

	if status, err := a.userSaver.CheckEmailIsAvailable(ctx, email); status != true || err != nil {
		if err != nil {
			log.Error("this email already taken", err)
		}

		return "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
	}

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	uid, err := middlewares.UUIDGenerator()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, uid, login, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			a.log.Warn("user already exists", err)

			return "", fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to save user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered", "id", id)

	return id, nil
}

func (a *Auth) Login(ctx context.Context, input, password string) (string, string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", input),
	)

	if status := middlewares.CheckLogin(input, password); status != true {
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("logging in")

	inputType := middlewares.IdentifyLoginInputType(input)

	user, err := a.userProvider.GetUser(ctx, inputType, input)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return "", "", fmt.Errorf("%s: %w", op, err)
		}

		a.log.Error("failed to get user", err)

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)

		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	accessToken, err := jwt.NewToken(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := middlewares.UUIDGenerator()
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in")

	return accessToken, refreshToken.String(), nil
}

//func (a *Auth) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
//	const op = "auth.RefreshToken"
//
//	log := a.log.With(
//		slog.String("op", op),
//		slog.String("refreshToken", refreshToken),
//	)
//
//	log.Info("refreshing tokens")
//
//	// Находим пользователя по refresh токену в Redis
//	userID, err := a.redisProvider.GetUserIDByRefreshToken(ctx, refreshToken) // Нужно реализовать этот метод
//	if err != nil {
//		a.log.Warn("invalid refresh token", err)
//		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
//	}
//
//	// Получаем данные пользователя
//	user, err := a.userProvider.GetUser(ctx, "id", userID)
//	if err != nil {
//		a.log.Error("failed to get user", err)
//		return "", "", fmt.Errorf("%s: %w", op, err)
//	}
//
//	// Генерируем новый access token
//	accessToken, err := jwt.NewToken(user, a.tokenTTL)
//	if err != nil {
//		a.log.Error("failed to generate access token", err)
//		return "", "", fmt.Errorf("%s: %w", op, err)
//	}
//
//	// Генерируем новый refresh token
//	newRefreshToken, err := middlewares.UUIDGenerator()
//	if err != nil {
//		a.log.Error("failed to generate refresh token", err)
//		return "", "", fmt.Errorf("%s: %w", op, err)
//	}
//
//	// Сохраняем новый refresh token в Redis и удаляем старый
//	err = a.redisProvider.SaveUserCache(ctx, newRefreshToken.String(), user.ID.String(), 7*24*time.Hour)
//	if err != nil {
//		a.log.Error("failed to save new refresh token", err)
//		return "", "", fmt.Errorf("%s: %w", op, err)
//	}
//
//	err = a.redisProvider.DeleteRefreshToken(ctx, refreshToken) // Удаление старого токена
//	if err != nil {
//		a.log.Error("failed to delete old refresh token", err)
//		return "", "", fmt.Errorf("%s: %w", op, err)
//	}
//
//	return accessToken, newRefreshToken.String(), nil
//}
