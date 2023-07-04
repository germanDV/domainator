package services

import (
	"github.com/go-playground/validator/v10"
)

// PingCreate is a struct that represents the payload sent by the user when creating a new ping
type PingCreate struct {
	Domain      string            `form:"domain" validate:"required,url"`
	SuccessCode int               `form:"success_code" validate:"required,gte=100,lte=599"`
	Errors      map[string]string `form:"-"`
}

// Validate makes PingCreate implement the Validator interface
func (pc *PingCreate) Validate(validate *validator.Validate) bool {
	err := validate.Struct(pc)

	if err != nil {
		pc.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				pc.Errors[e.Field()] = "This field is required"
				continue
			} else if tag == "gte" {
				pc.Errors[e.Field()] = "This field must be greater than or equal to 100"
				continue
			} else if tag == "lte" {
				pc.Errors[e.Field()] = "This field must be less than or equal to 599"
				continue
			} else if tag == "url" {
				pc.Errors[e.Field()] = "This field must be a valid URL"
			} else {
				pc.Errors[e.Field()] = e.Error()
				continue
			}
		}
	}

	return len(pc.Errors) == 0
}
