package certs

import (
	"regexp"
	"strings"
)

var hostnameRegexRFC952 = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`)

// TODO: rename `Domain` to `Hostname` everywhere
type Domain struct {
	value string
}

func ParseDomain(dom string) (Domain, error) {
	dom = strings.TrimSpace(dom)
	if dom == "" || !hostnameRegexRFC952.MatchString(dom) {
		return Domain{}, ErrInvalidDomain
	}
	return Domain{value: dom}, nil
}

func (d Domain) String() string {
	return string(d.value)
}
