package services

import (
	"github.com/go-playground/validator/v10"
)

// EmailUpdate is a struct that represents the payload sent to update email settings
type EmailUpdate struct {
	Email  string            `form:"email" validate:"required,email"`
	Errors map[string]string `form:"-"`
}

// Validate makes EmailUpdate implement the Validator interface
func (eu *EmailUpdate) Validate(validate *validator.Validate) bool {
	err := validate.Struct(eu)

	if err != nil {
		eu.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				eu.Errors[e.Field()] = "This field is required"
			} else if tag == "email" {
				eu.Errors[e.Field()] = "This field must be a valid email"
			} else {
				eu.Errors[e.Field()] = e.Error()
			}
		}
	}

	return len(eu.Errors) == 0
}
