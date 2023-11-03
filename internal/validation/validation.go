package validation

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// Validatable is an interface that structs that need to be validated must implement
type Validatable interface {
	Validate(validate *validator.Validate) bool
}

var (
	// ErrNotFound is an error that is returned when no records are found in the db
	ErrNotFound = errors.New("no records found")
	// ErrDuplicateDomain is an error that is returned when a domain is already in use
	ErrDuplicateDomain = errors.New("duplicate domain")
	// ErrDuplicateEmail is an error that is returned when an email is already in use
	ErrDuplicateEmail = errors.New("duplicate email")
	// ErrInvalidCode is an error that is returned when verification code is invalid
	ErrInvalidCode = errors.New("invalid or expired verification code")
)
