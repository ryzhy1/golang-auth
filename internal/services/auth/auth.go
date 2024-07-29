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
	userLogout   UserLogout
	userSaver    UserSaver
	userProvider UserProvider
	tokenTTL     time.Duration
}

type UserLogout interface {
	LogoutUser(ctx context.Context, login, token string) (status bool, err error)
	SaveUserCache(ctx context.Context, login, token string, duration time.Duration) error
}

type UserSaver interface {
	SaveUser(ctx context.Context, id uuid.UUID, login, email string, password []byte, createdAt time.Time) (uid string, err error)
}

type UserProvider interface {
	GetUser(ctx context.Context, input string) (user *models.User, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// New return a new instance of the Auth service
func New(log *slog.Logger, userLogout UserLogout, userSaver UserSaver, userProvider UserProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userLogout:   userLogout,
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

	id, err := a.userSaver.SaveUser(ctx, uid, login, email, passHash, time.Now())
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

func (a *Auth) Login(ctx context.Context, input, password string) (token string, err error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", input),
	)

	if status := middlewares.CheckLogin(input, password); status != true {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("logging in")

	user, err := a.userProvider.GetUser(ctx, input)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return "", fmt.Errorf("%s: %w", op, err)
		}

		a.log.Error("failed to get user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in")

	token, err = jwt.NewToken(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("adding token and user to cache")

	err = a.userLogout.SaveUserCache(ctx, user.Login, token, a.tokenTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) Logout(ctx context.Context, login, token string) (bool, error) {
	const op = "auth.Logout"

	log := a.log.With(
		slog.String("op", op),
		slog.String("login", login),
	)

	log.Info("logging out")

	status, err := a.userLogout.LogoutUser(ctx, login, token)
	if err != nil {
		if errors.Is(err, storage.ErrNoActiveSession) {
			a.log.Warn("user already logged out", err)

			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return status, nil
}
