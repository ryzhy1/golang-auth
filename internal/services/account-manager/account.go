package account_manager

import (
	"AuthService/internal/domain/models"
	"AuthService/internal/storage"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Account struct {
	log      *slog.Logger
	manager  Manager
	tokenTTL time.Duration
}

type Manager interface {
	CreatePurchase(ctx context.Context, username string, amount float64) error
	UpdateBalance(ctx context.Context, username string, amount float64) error
	GetUserByID(ctx context.Context, username string) (*models.User, error)
	UpdateEmail(ctx context.Context, username, newEmail string) error
	UpdatePassword(ctx context.Context, username string, newPassword []byte) error
}

var (
	ErrFailedToUpdate     = errors.New("failed to update")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyTaken  = errors.New("email already taken")
	ErrWrongEmail         = errors.New("wrong email")
	ErrWrongPassword      = errors.New("wrong password")
)

// New return a new instance of the Auth service
func New(log *slog.Logger, manager Manager, tokenTTL time.Duration) *Account {
	return &Account{
		log:      log,
		manager:  manager,
		tokenTTL: tokenTTL,
	}
}

func (a *Account) CreatePurchase(ctx context.Context, username string, amount float64) error {
	const op = "account.CreatePurchase"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", username),
	)

	user, err := a.manager.GetUserByID(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to update", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	userIDFromContext := ctx.Value("userID")
	if user.ID != userIDFromContext {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("creating purchase")

	if err := a.manager.CreatePurchase(ctx, username, amount); err != nil {
		log.Error("failed to create purchase", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("purchase created")

	return nil
}

func (a *Account) UpdateBalance(ctx context.Context, username string, amount float64) error {
	op := "account.UpdateBalance"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", username),
	)
	user, err := a.manager.GetUserByID(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to update", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	userIDFromContext := ctx.Value("userID")
	if user.ID != userIDFromContext {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("updating balance")

	if err := a.manager.UpdateBalance(ctx, username, amount); err != nil {
		log.Error("failed to update balance", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("balance updated")

	return nil
}

func (a *Account) GetUserByID(ctx context.Context, username string) (*models.User, error) {
	const op = "account.GetUserByID"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", username),
	)

	log.Info("searching user")

	user, err := a.manager.GetUserByID(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to get user", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userIDFromContext := ctx.Value("userID")
	if user.ID != userIDFromContext {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user found")

	return user, nil
}

func (a *Account) UpdateEmail(ctx context.Context, username, oldEmail, newEmail string) error {
	const op = "account.UpdateEmail"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", username),
	)

	user, err := a.manager.GetUserByID(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to update", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	userIDFromContext := ctx.Value("userID")
	if user.ID != userIDFromContext {
		return fmt.Errorf("%s: %w", op, err)
	}

	if user.Email != oldEmail {
		a.log.Warn("wrong email", err)
		return fmt.Errorf("%s: %w", op, storage.ErrWrongEmail)
	}

	log.Info("updating email")

	err = a.manager.UpdateEmail(ctx, username, newEmail)
	if err != nil {
		a.log.Error("failed to update", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("email updated")

	return nil
}

func (a *Account) UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	const op = "account.UpdatePassword"

	log := a.log.With(
		slog.String("op", op),
		slog.String("input", username),
	)

	user, err := a.manager.GetUserByID(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to update", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	userIDFromContext := ctx.Value("userID")
	if user.ID != userIDFromContext {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("updating password")

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(oldPassword)); err != nil {
		a.log.Info("invalid credentials", err)

		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)

		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.manager.UpdatePassword(ctx, username, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrWrongPassword) {
			a.log.Warn("wrong password", err)
			return fmt.Errorf("%s: %w", op, err)
		}

		a.log.Error("failed to update", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("password updated")

	return nil
}
