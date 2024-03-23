package certs

import (
	"time"
)

type RegisterReq struct {
	Domain Domain
	UserID ID
}

type GetAllReq struct {
	UserID ID
}

type UpdateReq struct {
	ID     ID
	UserID ID
}

type DeleteReq struct {
	ID     ID
	UserID ID
}

type Cert struct {
	ID        ID
	UserID    ID
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	Domain    Domain
	Issuer    Issuer
	Error     string
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

// serviceToRepoAdapter transforms a Cert from the Service layer to the Repository layer.
func serviceToRepoAdapter(cert Cert) repoCert {
	return repoCert{
		ID:        cert.ID.String(),
		UserID:    cert.UserID.String(),
		CreatedAt: cert.CreatedAt,
		UpdatedAt: cert.UpdatedAt,
		ExpiresAt: cert.ExpiresAt,
		Domain:    cert.Domain.String(),
		Issuer:    cert.Issuer.String(),
		Error:     cert.Error,
	}
}

// repoToServiceAdapter transforms a Cert from the Repository layer to the Service layer.
func repoToServiceAdapter(cert repoCert) (Cert, error) {
	parsedID, err := ParseID(cert.ID)
	if err != nil {
		return Cert{}, err
	}

	parsedUserID, err := ParseID(cert.UserID)
	if err != nil {
		return Cert{}, err
	}

	parsedDomain, err := ParseDomain(cert.Domain)
	if err != nil {
		return Cert{}, err
	}

	parsedIssuer, err := ParseIssuer(cert.Issuer)
	if err != nil {
		return Cert{}, err
	}

	return Cert{
		ID:        parsedID,
		UserID:    parsedUserID,
		CreatedAt: cert.CreatedAt,
		UpdatedAt: cert.UpdatedAt,
		ExpiresAt: cert.ExpiresAt,
		Domain:    parsedDomain,
		Issuer:    parsedIssuer,
		Error:     cert.Error,
	}, nil
}
