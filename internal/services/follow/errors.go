package follow

import "errors"

var (
	ErrFollowing    = errors.New("user is already following")
	ErrNoFollowing  = errors.New("user has not followed")
	ErrInvalidUUIDs = errors.New("some user does not exist")
)
