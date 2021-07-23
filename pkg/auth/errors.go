package auth

import "errors"

// TODO move auth into the internal directory

var (
	ErrInvalidCookie = errors.New("invalid cookie")
	ErrNotInCache    = errors.New("item not found in cache")
)
