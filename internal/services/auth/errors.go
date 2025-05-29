package auth

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid arguments")
	ErrExpired         = errors.New("expired token")
	ErrNoToken         = errors.New("no such token")
)
