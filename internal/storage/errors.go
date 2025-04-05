package storage

import "errors"

var (
	ErrNotFound         = errors.New("user is not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidArguments = errors.New("invalid arguments")
)
