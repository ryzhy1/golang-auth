package postgres

import (
	"AuthService/internal/domain/models"
	"AuthService/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"regexp"
	"time"
)

var (
	NullValue = int64(0)
)

type Storage struct {
	db *sqlx.DB
}

func New(connStr string) (*Storage, error) {
	const op = "storage.Postgres.New"

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, id uuid.UUID, login, email string, passHash []byte, created_at time.Time) (string, error) {
	const op = "storage.Postgres.SaveUser"

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

	err = stmt.QueryRowContext(ctx, id, login, email, passHash, created_at).Scan(&id)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id.String(), nil
}

// IdentifyInputType determines whether the input is an email or login
func IdentifyInputType(input string) string {
	const emailPattern = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	emailRegex := regexp.MustCompile(emailPattern)

	if emailRegex.MatchString(input) {
		return "email"
	}
	return "login"
}

// GetUser fetches a user by login or email
func (s *Storage) GetUser(ctx context.Context, input string) (*models.User, error) {
	const op = "storage.Postgres.GetUser"

	var user models.User
	var query string
	var err error

	switch IdentifyInputType(input) {
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

func (s *Storage) Close() error {
	return s.db.Close()
}
