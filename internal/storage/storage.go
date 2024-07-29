package storage

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNoActiveSession   = errors.New("user already logged out")
)
