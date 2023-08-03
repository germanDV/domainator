package certs

import (
	"domainator/internal/certstatus"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Cert represents the information for a TLS certificate.
type Cert struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Domain    string
	CreatedAt time.Time
}

// Summary represents the information for a TLS certificate including its latest check.
type Summary struct {
	ID        uuid.UUID
	Domain    string
	Status    certstatus.Status
	Expiry    time.Time
	LastCheck time.Time
}

// Check represents the result of a TLS certificate check.
type Check struct {
	ID         uuid.UUID
	CertID     uuid.UUID
	RespStatus certstatus.Status
	Expiry     time.Time
	CreatedAt  time.Time
}

// CreateCertReq represents the request to save a new certificate for checking.
type CreateCertReq struct {
	Domain string            `form:"domain" validate:"required,hostname"`
	Errors map[string]string `form:"-"`
}

// Validate makes CreateCertReq implement the validation.Validatable interface.
func (ccr *CreateCertReq) Validate(validate *validator.Validate) bool {
	err := validate.Struct(ccr)

	if err != nil {
		ccr.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				ccr.Errors[e.Field()] = "This field is required"
			} else if tag == "hostname" {
				ccr.Errors[e.Field()] = "This field must be a valid hostname (e.g. example.com)"
			} else {
				ccr.Errors[e.Field()] = e.Error()
			}
		}
	}

	return len(ccr.Errors) == 0
}
