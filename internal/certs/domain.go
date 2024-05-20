package certs

import (
	"fmt"
	"regexp"
	"strings"
)

var hostnameRegexRFC952 = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`)

type Domain struct {
	value string
}

func ParseDomain(dom string) (Domain, error) {
	dom = strings.TrimSpace(dom)
	if dom == "" || !hostnameRegexRFC952.MatchString(dom) {
		return Domain{}, fmt.Errorf("error parsing domain %s: %w", dom, ErrInvalidDomain)
	}
	return Domain{value: dom}, nil
}

func (d Domain) String() string {
	return d.value
}
