package storage

import "errors"

var (
	ErrNotFound       = errors.New("user is not found")
	ErrUserExists     = errors.New("user already exists")
	ErrInvalidUserKey = errors.New("unknown type of user key")
)
