package services

import (
	"github.com/go-playground/validator/v10"
)

// UserCredentials is a struct that represents the payload sent to sign up / in
type UserCredentials struct {
	Email    string            `form:"email" validate:"required,email"`
	Password string            `form:"password" validate:"required,alphanum,gte=8"`
	Errors   map[string]string `form:"-"`
}

// Validate makes UserCreate implement the Validator interface
func (uc *UserCredentials) Validate(validate *validator.Validate) bool {
	err := validate.Struct(uc)

	if err != nil {
		uc.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				uc.Errors[e.Field()] = "This field is required"
				continue
			} else if tag == "gte" {
				uc.Errors[e.Field()] = "This field must be greater than or equal to 8"
				continue
			} else if tag == "email" {
				uc.Errors[e.Field()] = "This field must be a valid email"
			} else if tag == "alphanum" {
				uc.Errors[e.Field()] = "This field must be alphanumeric"
			} else {
				uc.Errors[e.Field()] = e.Error()
				continue
			}
		}
	}

	return len(uc.Errors) == 0
}
