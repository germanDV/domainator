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
	ErrInvalidIssuer   = errors.New("issuer is required")
)

var hostnameRegexRFC952 = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`)

type ID string

type Domain string

type Issuer string

// Cert is an aggregate root that represents the information for a TLS certificate.
type Cert struct {
	ID        ID
	CreatedAt time.Time
	ExpiresAt time.Time
	Domain    Domain
	Issuer    Issuer
}

func New(domain Domain, issuer Issuer, expiresAt time.Time) Cert {
	return Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Domain:    domain,
		Issuer:    issuer,
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

func NewIssuer(issuer string) (Issuer, error) {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return Issuer(""), ErrInvalidIssuer
	}
	return Issuer(issuer), nil
}

func (i Issuer) String() string {
	return string(i)
}
