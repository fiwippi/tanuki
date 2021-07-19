package auth

import "errors"

var (
	ErrInvalidCookie = errors.New("invalid cookie")
	ErrNotInCache = errors.New("item not found in cache")
)

