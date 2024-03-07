package certs

import (
	"fmt"
	"time"

	"github.com/germandv/domainator/internal/tlser"
)

type Service interface {
	RegisterCert(dto RegisterCertReq) (CertDto, error)
	GetAll() ([]CertDto, error)
	Delete(id string) error
	Update(id string) (CertDto, error)
}

type CertsService struct {
	repo      Repo
	tlsClient tlser.Client
}

func NewService(tlsClient tlser.Client) *CertsService {
	repo := NewRepo()

	// Add some dummy data
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 6, 0),
		Domain:    Domain("go.dev"),
		Issuer:    Issuer("Let's Encrypt"),
		Error:     "",
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, -2, 0),
		Domain:    Domain("archlinux.org"),
		Issuer:    Issuer("Certigo"),
		Error:     "",
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 10),
		Domain:    Domain("debian.org"),
		Issuer:    Issuer("Comodo"),
		Error:     "",
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 10),
		Domain:    Domain("yahoo.xyz"),
		Issuer:    Issuer("Comodo"),
		Error:     "Could not connect",
	})

	return &CertsService{
		repo:      repo,
		tlsClient: tlsClient,
	}
}

func (s *CertsService) RegisterCert(dto RegisterCertReq) (CertDto, error) {
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

	cert := New(domain, issuer, data.Expiry)
	err = s.repo.Save(cert)
	if err != nil {
		return CertDto{}, err
	}

	return toDTO(cert), nil
}

func (s *CertsService) GetAll() ([]CertDto, error) {
	certificates, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	dtos := []CertDto{}
	for _, c := range certificates {
		dtos = append(dtos, toDTO(c))
	}

	return dtos, nil
}

func (s *CertsService) Delete(id string) error {
	parsedID, err := ParseID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(parsedID)
}

func (s *CertsService) Update(id string) (CertDto, error) {
	parsedID, err := ParseID(id)
	if err != nil {
		return CertDto{}, err
	}

	cert, err := s.repo.Get(parsedID)
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

	c, err := s.repo.Update(parsedID, data.Expiry, issuer, e)
	if err != nil {
		return CertDto{}, err
	}

	return toDTO(c), nil
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
