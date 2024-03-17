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
	ErrNotFound        = errors.New("not found")
)

var hostnameRegexRFC952 = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`)

type ID string

type Domain string

type Issuer string

// Cert is an aggregate root that represents the information for a TLS certificate.
type Cert struct {
	ID        ID        `db:"id"`
	UserID    ID        `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Domain    Domain    `db:"domain"`
	Issuer    Issuer    `db:"issuer"`
	Error     string    `db:"error"`
}

func New(userID ID, domain Domain, issuer Issuer, expiresAt time.Time) Cert {
	return Cert{
		ID:        NewID(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Domain:    domain,
		Issuer:    issuer,
		Error:     "",
	}
}

func NewID() ID {
	id, _ := uuid.NewV7()
	return ID(id.String())
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
