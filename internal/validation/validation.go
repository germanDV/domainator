// Package validation provides some utilities regarding validation and errors.
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
	ErrNotFound = errors.New("No records found")
	// ErrDuplicateDomain is an error that is returned when a domain is already in use
	ErrDuplicateDomain = errors.New("Duplicate domain")
	// ErrDuplicateEmail is an error that is returned when an email is already in use
	ErrDuplicateEmail = errors.New("Duplicate email")
	// ErrInvalidCode is an error that is returned when verification code is invalid
	ErrInvalidCode = errors.New("Invalid or expired verification code")
)
