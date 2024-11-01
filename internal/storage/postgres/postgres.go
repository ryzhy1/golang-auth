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

func (s *Storage) SaveUser(ctx context.Context, id uuid.UUID, username, email string, passHash []byte) (string, error) {
	const op = "storage.Postgres.SaveUser"

	sql, args, err := squirrel.Insert("users").
		Columns("id", "username", "email", "password", "created_at").
		Values(id, username, email, passHash, time.Now()).
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

func (s *Storage) CheckUserByEmail(ctx context.Context, userId, email string) error {
	const op = "storage.Postgres.GetUserEmail"

	sql, args, err := squirrel.Select("email").
		From("users").
		Where(squirrel.Eq{"id": userId, "password": email}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var userEmail string
	err = s.db.QueryRow(ctx, sql, args...).Scan(&userEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, storage.ErrWrongEmail)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateEmail(ctx context.Context, userId, email string) error {
	const op = "storage.Postgres.UpdateEmail"

	sql, args, err := squirrel.Update("users").
		SetMap(squirrel.Eq{"email": email}).
		Where(squirrel.Eq{"id": userId}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CheckUserByPassword(ctx context.Context, userId, password string) (string, error) {
	const op = "storage.Postgres.CheckUserByPassword"

	sql, args, err := squirrel.Select("password").
		From("users").
		Where(squirrel.Eq{"id": userId}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var userPassword string
	err = s.db.QueryRow(ctx, sql, args...).Scan(&userPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrWrongPassword)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userPassword, nil
}

func (s *Storage) UpdatePassword(ctx context.Context, userId, password string) error {
	const op = "storage.Postgres.UpdatePassword"

	sql, args, err := squirrel.Update("users").
		SetMap(squirrel.Eq{"password": password}).
		Where(squirrel.Eq{"id": userId}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}
