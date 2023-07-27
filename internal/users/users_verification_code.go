package users

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// VerificationCode is a struct that represents the payload sent to verify a user
type VerificationCode struct {
	Email     string            `form:"email" validate:"required,email"`
	Plain     string            `form:"code" validate:"required,len=9"`
	Hash      []byte            `form:"-"`
	ExpiresAt time.Time         `form:"-"`
	Errors    map[string]string `form:"-"`
}

// Validate makes VerificationCode implement the Validator interface
func (vc *VerificationCode) Validate(validate *validator.Validate) bool {
	err := validate.Struct(vc)

	if err != nil {
		vc.Errors = make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			tag := e.Tag()
			if tag == "required" {
				vc.Errors[e.Field()] = "This field is required"
			} else if tag == "len" {
				vc.Errors[e.Field()] = "This field must be 9 characters long"
			} else if tag == "email" {
				vc.Errors[e.Field()] = "This field must be a valid email address"
			} else {
				vc.Errors[e.Field()] = e.Error()
			}
		}
	}

	return len(vc.Errors) == 0
}

// Matches checks whether the provided code matches the hashed verification code.
func (vc *VerificationCode) Matches(candidate string) bool {
	err := bcrypt.CompareHashAndPassword(vc.Hash, []byte(candidate))
	return err == nil
}

// generateCode generates a random code with format xxxx-xxxx (e.g. 1234-5678).
// It panics if it fails to generate a random number.
func generateCode() string {
	max := big.NewInt(10000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	m, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%04d-%04d", n, m)
}

// hashCode calculates the bcrypt hash of a code.
// It panics if it fails to generate the hash.
func hashCode(code string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}
