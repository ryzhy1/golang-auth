package storage

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNoActiveSession   = errors.New("user already logged out")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrWrongEmail        = errors.New("wrong email")
	ErrWrongPassword     = errors.New("wrong password")
)
