package postgres

import (
	"AuthService/internal/domain/models"
	"AuthService/internal/storage"
	"AuthService/middlewares"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type Storage struct {
	db *sqlx.DB
}

func New(connStr string) (*Storage, error) {
	const op = "storage.DB.New"

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, id uuid.UUID, login, email string, passHash []byte, createdAt time.Time) (string, error) {
	const op = "storage.Postgres.SaveUser"

	if u, _ := s.GetUser(ctx, login); u != nil {
		return "", fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
	}

	if u, _ := s.GetUser(ctx, email); u != nil {
		return "", fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
	}

	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (id, login, email, password, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(stmt)

	err = stmt.QueryRowContext(ctx, id, login, email, passHash, createdAt).Scan(&id)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id.String(), nil
}

// GetUser fetches a user by login or email
func (s *Storage) GetUser(ctx context.Context, input string) (*models.User, error) {
	const op = "storage.Postgres.GetUser"

	var user models.User
	var query string
	var err error

	switch middlewares.IdentifyLoginInputType(input) {
	case "email":
		query = "SELECT id, login, email, password FROM users WHERE email = $1"
		err = s.db.GetContext(ctx, &user, query, input)
	case "login":
		query = "SELECT id, login, email, password FROM users WHERE login = $1"
		err = s.db.GetContext(ctx, &user, query, input)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) CreatePurchase(ctx context.Context, userID string, amount float64) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) UpdateBalance(ctx context.Context, username string, amount float64) error {
	const op = "storage.Postgres.UpdateBalance"

	query := "UPDATE users SET balance = $1, updated_at = NOW() WHERE login = $2"
	_, err := s.db.ExecContext(ctx, query, amount, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserByID(ctx context.Context, username string) (*models.User, error) {
	const op = "storage.Postgres.GetUserByID"

	var user models.User
	query := "SELECT id, login, email, password, balance, discount FROM users WHERE login = $1"
	err := s.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) UpdateEmail(ctx context.Context, username, newEmail string) error {
	const op = "storage.Postgres.UpdateEmail"

	var user models.User

	query := "SELECT id FROM users WHERE email = $1"
	err := s.db.GetContext(ctx, &user, query, newEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, storage.ErrEmailAlreadyTaken)
	}

	query = "UPDATE users SET email = $1, updated_at = NOW() WHERE login = $2"
	_, err = s.db.ExecContext(ctx, query, newEmail, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdatePassword(ctx context.Context, username string, newPassword []byte) error {
	const op = "storage.Postgres.UpdatePassword"

	query := "UPDATE users SET password = $1, updated_at = NOW() WHERE login = $2"
	_, err := s.db.ExecContext(ctx, query, newPassword, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
