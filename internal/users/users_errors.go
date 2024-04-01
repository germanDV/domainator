package users

import "errors"

var (
	ErrInvalidEmail   = errors.New("email is required and must be a valid email address")
	ErrDuplicateEmail = errors.New("email already exists")
	ErrNotFound       = errors.New("user not found")
)
