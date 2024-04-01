package certs

import "strings"

type Issuer struct {
	value string
}

func ParseIssuer(issuer string) (Issuer, error) {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return Issuer{}, ErrInvalidIssuer
	}
	return Issuer{value: issuer}, nil
}

func (i Issuer) String() string {
	return i.value
}
