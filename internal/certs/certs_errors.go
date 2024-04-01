package certs

import "errors"

var (
	ErrInvalidDomain   = errors.New("domain is required and must be a valid hostname")
	ErrDuplicateDomain = errors.New("domain already exists")
	ErrInvalidIssuer   = errors.New("issuer is required")
	ErrNotFound        = errors.New("domain not found")
)
