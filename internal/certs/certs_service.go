package certs

import (
	"context"
	"fmt"
	"time"

	"github.com/germandv/domainator/internal/tlser"
)

type Service interface {
	Save(context.Context, RegisterReq) (Cert, error)
	GetAll(context.Context, GetAllReq) ([]Cert, error)
	Delete(context.Context, DeleteReq) error
	Update(context.Context, UpdateReq) (Cert, error)
}

type CertsService struct {
	repo      Repo
	tlsClient tlser.Client
}

func NewService(tlsClient tlser.Client, repo Repo) *CertsService {
	return &CertsService{
		repo:      repo,
		tlsClient: tlsClient,
	}
}

func (s *CertsService) Save(ctx context.Context, req RegisterReq) (Cert, error) {
	data := s.tlsClient.GetCertData(req.Domain.value)
	if data.Status != tlser.StatusOK {
		return Cert{}, fmt.Errorf("TLS error: %s", data.Status)
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return Cert{}, err
	}

	cert := New(req.UserID, req.Domain, issuer, data.Expiry)
	err = s.repo.Save(ctx, serviceToRepoAdapter(cert))
	if err != nil {
		return Cert{}, err
	}

	return cert, nil
}

func (s *CertsService) GetAll(ctx context.Context, req GetAllReq) ([]Cert, error) {
	certificates, err := s.repo.GetAll(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	certs := make([]Cert, len(certificates))
	for i, c := range certificates {
		cert, err := repoToServiceAdapter(c)
		if err != nil {
			return nil, err
		}
		certs[i] = cert
	}

	return certs, nil
}

func (s *CertsService) Delete(ctx context.Context, req DeleteReq) error {
	return s.repo.Delete(ctx, req.UserID, req.ID)
}

func (s *CertsService) Update(ctx context.Context, req UpdateReq) (Cert, error) {
	cert, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return Cert{}, err
	}

	e := ""
	data := s.tlsClient.GetCertData(cert.Domain)
	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		e = string(data.Status)
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return Cert{}, ErrInvalidIssuer
	}

	now := time.Now().UTC()
	err = s.repo.Update(ctx, req.UserID, req.ID, data.Expiry, issuer.value, now, e)
	if err != nil {
		return Cert{}, err
	}

	cert.UpdatedAt = now
	cert.Issuer = issuer.value
	cert.ExpiresAt = data.Expiry
	cert.Error = e

	c, err := repoToServiceAdapter(cert)
	if err != nil {
		return Cert{}, err
	}
	return c, nil
}
