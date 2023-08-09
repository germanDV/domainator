package events

import (
	"github.com/google/uuid"

	"github.com/go-playground/validator/v10"
)

// Event represents an application event that we want to store in the database.
type Event struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Name    string
	Payload map[string]any
}

// CreateEventReq represents the request to create an event.
type CreateEventReq struct {
	Name    string            `form:"name" validate:"required,lte=64"`
	Payload map[string]any    `form:"payload"`
	Errors  map[string]string `form:"-"`
}

// Validate makes EventCreateReq implement the validation.Validatable interface.
func (ecr *CreateEventReq) Validate(validate *validator.Validate) bool {
	err := validate.Struct(ecr)

	if err != nil {
		ecr.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				ecr.Errors[e.Field()] = "This field is required"
			} else if tag == "lte" {
				ecr.Errors[e.Field()] = "This field must have at most 64 characters"
			} else {
				ecr.Errors[e.Field()] = e.Error()
			}
		}
	}

	return len(ecr.Errors) == 0
}
