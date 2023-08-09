package endpoints

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Summary represents a ping's settings and its latest check.
type Summary struct {
	ID        uuid.UUID
	Domain    string
	Status    string // 'healthy' or 'unhealthy'
	LastCheck time.Time
}

// CreateEndpointReq is a struct that represents the payload sent by the user when creating a new Endpoint.
type CreateEndpointReq struct {
	Domain      string            `form:"domain" validate:"required,url"`
	SuccessCode int               `form:"success_code" validate:"required,gte=100,lte=599"`
	Errors      map[string]string `form:"-"`
}

// Validate makes CreateEndpointReq implement the validation.Validatable interface.
func (cp *CreateEndpointReq) Validate(validate *validator.Validate) bool {
	err := validate.Struct(cp)

	if err != nil {
		cp.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				cp.Errors[e.Field()] = "This field is required"
			} else if tag == "gte" {
				cp.Errors[e.Field()] = "This field must be greater than or equal to 100"
			} else if tag == "lte" {
				cp.Errors[e.Field()] = "This field must be less than or equal to 599"
			} else if tag == "url" {
				cp.Errors[e.Field()] = "This field must be a valid URL"
			} else {
				cp.Errors[e.Field()] = e.Error()
			}
		}
	}

	return len(cp.Errors) == 0
}

// Endpoint represents a domain to ping.
type Endpoint struct {
	ID          uuid.UUID
	Domain      string
	SuccessCode int
	CreatedAt   time.Time
}

// Healthcheck represents a check ("ping") to an Endpoint.
type Healthcheck struct {
	ID         uuid.UUID
	EndpointID uuid.UUID
	RespStatus int
	TookMs     int
	CreatedAt  time.Time
}
