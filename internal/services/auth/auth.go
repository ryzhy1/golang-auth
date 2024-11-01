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
	log            *slog.Logger
	userRepository UserRepository
	tokenTTL       time.Duration
}

type UserRepository interface {
	SaveUser(ctx context.Context, id uuid.UUID, login, email string, password []byte) (uid string, err error)
	GetUser(ctx context.Context, inputType, input string) (user *models.User, err error)
	CheckUsernameIsAvailable(ctx context.Context, login string) (status bool, err error)
	CheckEmailIsAvailable(ctx context.Context, email string) (status bool, err error)
	CheckUserByEmail(ctx context.Context, userId, email string) error
	CheckUserByPassword(ctx context.Context, userId, password string) (hashPassword string, err error)
	UpdateEmail(ctx context.Context, userId, email string) error
	UpdatePassword(ctx context.Context, userId, password string) error
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// New return a new instance of the Auth service
func New(log *slog.Logger, userRepository UserRepository, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:            log,
		userRepository: userRepository,
		tokenTTL:       tokenTTL,
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

	if status, err := a.userRepository.CheckUsernameIsAvailable(ctx, login); status != true || err != nil {
		if err != nil {
			log.Error("this username already taken", err)
		}

		return "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
	}

	if status, err := a.userRepository.CheckEmailIsAvailable(ctx, email); status != true || err != nil {
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

	id, err := a.userRepository.SaveUser(ctx, uid, login, email, passHash)
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

	user, err := a.userRepository.GetUser(ctx, inputType, input)
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

func (a *Auth) UpdateUserEmail(ctx context.Context, userId, oldEmail, newEmail string) error {
	const op = "auth.GetUserEmail"

	log := a.log.With(
		slog.String("op", op),
		slog.String("userId", userId),
	)

	log.Info("getting user email")

	if status := middlewares.CorrectEmailChecker(oldEmail); status != true {
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if status := middlewares.CorrectEmailChecker(newEmail); status != true {
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	err := a.userRepository.CheckUserByEmail(ctx, userId, oldEmail)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return fmt.Errorf("%s: %w", op, err)
		}

		a.log.Error("failed to get user", err)

		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.userRepository.UpdateEmail(ctx, userId, newEmail)
	if err != nil {
		a.log.Error("failed to update user email", err)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) UpdateUserPassword(ctx context.Context, userId, oldPassword, newPassword string) error {
	const op = "auth.UpdateUserPassword"

	log := a.log.With(
		slog.String("op", op),
		slog.String("userId", userId),
	)

	log.Info("checking user credentials")

	if len(oldPassword) < 8 || len(newPassword) < 8 {
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if oldPassword == newPassword {
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("picking user password from database")

	password, err := a.userRepository.CheckUserByPassword(ctx, userId, oldPassword)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return fmt.Errorf("%s: %w", op, err)
		}

		a.log.Error("failed to get user", err)

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("comparing users password")

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(oldPassword))
	if err != nil {
		a.log.Info("invalid credentials", err)

		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("updating user password")

	err = a.userRepository.UpdatePassword(ctx, userId, newPassword)
	if err != nil {
		a.log.Error("failed to update user password", err)

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("password updated successfully")

	return nil
}
