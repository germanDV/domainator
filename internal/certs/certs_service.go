package certs

import (
	"context"
	"fmt"
	"time"

	"github.com/germandv/domainator/internal/tlser"
)

type Service interface {
	Save(context.Context, RegisterCertReq) (CertDto, error)
	GetAll(context.Context, GetAllCertsReq) ([]CertDto, error)
	Delete(context.Context, DeleteCertReq) error
	Update(context.Context, UpdateCertReq) (CertDto, error)
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

func (s *CertsService) Save(ctx context.Context, dto RegisterCertReq) (CertDto, error) {
	userID, err := ParseID(dto.UserID)
	if err != nil {
		return CertDto{}, err
	}

	domain, err := NewDomain(dto.Domain)
	if err != nil {
		return CertDto{}, err
	}

	data := s.tlsClient.GetCertData(domain.String())
	if data.Status != tlser.StatusOK {
		return CertDto{}, fmt.Errorf("TLS error: %s", data.Status)
	}

	issuer, err := NewIssuer(data.Issuer)
	if err != nil {
		return CertDto{}, err
	}

	cert := New(userID, domain, issuer, data.Expiry)
	err = s.repo.Save(ctx, cert)
	if err != nil {
		return CertDto{}, err
	}

	return toDTO(cert), nil
}

func (s *CertsService) GetAll(ctx context.Context, dto GetAllCertsReq) ([]CertDto, error) {
	userID, err := ParseID(dto.UserID)
	if err != nil {
		return nil, err
	}

	certificates, err := s.repo.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := []CertDto{}
	for _, c := range certificates {
		dtos = append(dtos, toDTO(c))
	}

	return dtos, nil
}

func (s *CertsService) Delete(ctx context.Context, dto DeleteCertReq) error {
	parsedID, err := ParseID(dto.UserID)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, parsedID)
}

func (s *CertsService) Update(ctx context.Context, dto UpdateCertReq) (CertDto, error) {
	parsedID, err := ParseID(dto.UserID)
	if err != nil {
		return CertDto{}, err
	}

	cert, err := s.repo.Get(ctx, parsedID)
	if err != nil {
		return CertDto{}, err
	}

	e := ""
	data := s.tlsClient.GetCertData(cert.Domain.String())
	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		e = string(data.Status)
	}

	issuer, err := NewIssuer(data.Issuer)
	if err != nil {
		return CertDto{}, ErrInvalidIssuer
	}

	now := time.Now().UTC()
	err = s.repo.Update(ctx, parsedID, data.Expiry, issuer, now, e)
	if err != nil {
		return CertDto{}, err
	}

	cert.UpdatedAt = now
	cert.Issuer = issuer
	cert.ExpiresAt = data.Expiry
	cert.Error = e

	return toDTO(cert), nil
}

func toDTO(c Cert) CertDto {
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

	return CertDto{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt.Format(time.DateOnly),
		ExpiresAt: c.ExpiresAt.Format(time.DateOnly),
		Domain:    c.Domain.String(),
		Issuer:    c.Issuer.String(),
		Status:    status,
		Error:     c.Error,
	}
}
