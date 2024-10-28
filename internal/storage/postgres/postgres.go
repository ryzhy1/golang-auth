package postgres

import (
	"AuthService/internal/domain/models"
	"AuthService/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	_ "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewPostgres(conn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) SaveUser(ctx context.Context, id uuid.UUID, username, email string, passHash []byte,
	createdAt time.Time) (string, error) {
	const op = "storage.Postgres.SaveUser"

	sql, args, err := squirrel.Insert("users").
		Columns("id", "username", "email", "password", "created_at").
		Values(id, username, email, passHash, createdAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id.String(), nil
}

// GetUser fetches a user by login or email
func (s *Storage) GetUser(ctx context.Context, inputType, input string) (*models.User, error) {
	const op = "storage.Postgres.GetUser"

	var pgUUID pgtype.UUID
	var user models.User

	sql, args, err := squirrel.Select("id", "username", "email", "password").
		From("users").
		Where(squirrel.Eq{inputType: input}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.db.QueryRow(ctx, sql, args...).Scan(&pgUUID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.ID = uuid.UUID(pgUUID.Bytes)

	return &user, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}

func (s *Storage) CheckUsernameIsAvailable(ctx context.Context, input string) (bool, error) {
	const op = "storage.CheckLoginIsAvailable"

	sql, args, err := squirrel.Select("id").
		From("users").
		Where(squirrel.Eq{"username": input}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = s.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return true, nil // Логин доступен
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return false, nil
}

func (s *Storage) CheckEmailIsAvailable(ctx context.Context, email string) (bool, error) {
	const op = "storage.CheckEmailIsAvailable"

	sql, args, err := squirrel.Select("id").
		From("users").
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = s.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return true, nil // Email доступен
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return false, nil
}
