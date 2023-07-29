package users

import (
	"time"

	"domainator/internal/notificators"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// User is a struct that represents a user
type User struct {
	ID        uuid.UUID `form:"id"`
	Email     string    `form:"email"`
	Password  pwd       `form:"-"`
	Activated bool      `form:"activated"`
	CreatedAt time.Time `form:"created_at"`
}

// newUser returns a User struct, hashing the password.
func newUser(email, password string) (*User, error) {
	hashedPwd, err := hashPwd(password, 12)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:        uuid.New(),
		Email:     email,
		Activated: false,
		CreatedAt: time.Now().UTC(),
		Password: pwd{
			plain: &password,
			hash:  hashedPwd,
		},
	}

	return user, nil
}

// UserCredentials is a struct that represents the payload sent to sign up / in
type UserCredentials struct {
	Email    string            `form:"email" validate:"required,email"`
	Password string            `form:"password" validate:"required,alphanum,gte=8"`
	Errors   map[string]string `form:"-"`
}

// Validate makes UserCreate implement the validation.Validatable interface
func (uc *UserCredentials) Validate(validate *validator.Validate) bool {
	err := validate.Struct(uc)

	if err != nil {
		uc.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				uc.Errors[e.Field()] = "This field is required"
			} else if tag == "gte" {
				uc.Errors[e.Field()] = "This field must be greater than or equal to 8"
			} else if tag == "email" {
				uc.Errors[e.Field()] = "This field must be a valid email"
			} else if tag == "alphanum" {
				uc.Errors[e.Field()] = "This field must be alphanumeric"
			} else {
				uc.Errors[e.Field()] = e.Error()
			}
		}
	}

	return len(uc.Errors) == 0
}

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

// NotificationPref is a struct that represents a user's notification preference
type NotificationPref struct {
	ID      int
	Service notificators.Service
	Enabled bool
	To      string // email address | slack webhook url
}
