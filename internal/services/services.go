// Package services is a package that encapsulates the services used by the appliation.
package services

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
	ErrNotFound = errors.New("services: no records found")
	// ErrDuplicateDomain is an error that is returned when a domain is already in use
	ErrDuplicateDomain = errors.New("services: duplicate domain")
	// ErrDuplicateEmail is an error that is returned when an email is already in use
	ErrDuplicateEmail = errors.New("services: duplicate email")
)
