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
	ErrNotFound        = errors.New("services: no records found")
	ErrDuplicateDomain = errors.New("services: duplicate domain")
)
