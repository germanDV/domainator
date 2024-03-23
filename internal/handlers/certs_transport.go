package handlers

import (
	"fmt"
	"time"

	"github.com/germandv/domainator/internal/certs"
)

type RegisterCertReq struct {
	Domain string
	UserID string
}

func (r RegisterCertReq) Parse() (certs.RegisterReq, error) {
	domain, err := certs.ParseDomain(r.Domain)
	if err != nil {
		return certs.RegisterReq{}, err
	}

	userID, err := certs.ParseID(r.UserID)
	if err != nil {
		return certs.RegisterReq{}, err
	}

	return certs.RegisterReq{
		Domain: domain,
		UserID: userID,
	}, nil
}

type GetAllCertsReq struct {
	UserID string
}

func (r GetAllCertsReq) Parse() (certs.GetAllReq, error) {
	userID, err := certs.ParseID(r.UserID)
	if err != nil {
		return certs.GetAllReq{}, err
	}

	return certs.GetAllReq{
		UserID: userID,
	}, nil
}

type UpdateCertReq struct {
	ID     string
	UserID string
}

func (r UpdateCertReq) Parse() (certs.UpdateReq, error) {
	id, err := certs.ParseID(r.ID)
	if err != nil {
		return certs.UpdateReq{}, err
	}

	userID, err := certs.ParseID(r.UserID)
	if err != nil {
		return certs.UpdateReq{}, err
	}

	return certs.UpdateReq{
		ID:     id,
		UserID: userID,
	}, nil
}

type DeleteCertReq struct {
	ID     string
	UserID string
}

func (r DeleteCertReq) Parse() (certs.DeleteReq, error) {
	id, err := certs.ParseID(r.ID)
	if err != nil {
		return certs.DeleteReq{}, err
	}

	userID, err := certs.ParseID(r.UserID)
	if err != nil {
		return certs.DeleteReq{}, err
	}

	return certs.DeleteReq{
		ID:     id,
		UserID: userID,
	}, nil
}

type TransportCert struct {
	ID        string
	CreatedAt string
	ExpiresAt string
	Domain    string
	Issuer    string
	Status    string
	Error     string
}

// serviceToTransportAdapter transforms a Cert from the Service layer to the Transport layer.
func serviceToTransportAdapter(c certs.Cert) TransportCert {
	now := time.Now()
	diffDays := c.ExpiresAt.Sub(now).Hours() / 24
	status := ""
	if c.Error != "" {
		status = c.Error
	} else if diffDays < 0 {
		status = "Expired"
	} else {
		status = fmt.Sprintf("Expires in %d days", int(diffDays))
	}

	return TransportCert{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt.Format(time.DateOnly),
		ExpiresAt: c.ExpiresAt.Format(time.DateOnly),
		Domain:    c.Domain.String(),
		Issuer:    c.Issuer.String(),
		Status:    status,
		Error:     c.Error,
	}
}
