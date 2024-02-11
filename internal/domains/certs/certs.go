package certs

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDomain   = errors.New("domain is required and must be a valid hostname")
	ErrDuplicateDomain = errors.New("domain already exists")
)

var hostnameRegexRFC952 = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`)

type ID string

type Domain string

// Cert is an aggregate root that represents the information for a TLS certificate.
type Cert struct {
	ID        ID
	CreatedAt time.Time
	Domain    Domain
}

func New(domain Domain) Cert {
	return Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		Domain:    domain,
	}
}

func NewID() ID {
	return ID(uuid.NewString())
}

func ParseID(id string) (ID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ID(""), err
	}
	return ID(parsedID.String()), nil
}

func (id ID) String() string {
	return string(id)
}

func NewDomain(dom string) (Domain, error) {
	dom = strings.TrimSpace(dom)
	if dom == "" || !hostnameRegexRFC952.MatchString(dom) {
		return Domain(""), ErrInvalidDomain
	}
	return Domain(dom), nil
}

func (d Domain) String() string {
	return string(d)
}
