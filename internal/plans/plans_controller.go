// Package plans holds the logic for the plans service.
package plans

import "github.com/go-playground/validator/v10"

// Controller is a controller that handles requests to the plans service.
type Controller struct {
	repo      Repo
	validator *validator.Validate
}

// NewController returns a new plans controller.
func NewController(repo Repo, validate *validator.Validate) *Controller {
	return &Controller{
		repo:      repo,
		validator: validate,
	}
}
