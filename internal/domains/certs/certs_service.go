package certs

import "time"

type Service interface {
	RegisterCert(dto RegisterCertReq) (CertDto, error)
	GetAll() ([]CertDto, error)
	Delete(id string) error
}

type CertsService struct {
	repo Repo
}

func NewService() *CertsService {
	repo := NewRepo()

	// Add some dummy data
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 6, 0),
		Domain:    Domain("go.dev"),
		Issuer:    Issuer("Let's Encrypt"),
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, -2, 0),
		Domain:    Domain("archlinux.org"),
		Issuer:    Issuer("Certigo"),
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 10),
		Domain:    Domain("debian.org"),
		Issuer:    Issuer("Comodo"),
	})

	return &CertsService{
		repo: repo,
	}
}

func (s *CertsService) RegisterCert(dto RegisterCertReq) (CertDto, error) {
	domain, err := NewDomain(dto.Domain)
	if err != nil {
		return CertDto{}, err
	}

	// TODO: do this for real
	issuer, err := NewIssuer("Let's Encrypt")
	if err != nil {
		return CertDto{}, err
	}

	expiresAt := time.Now().AddDate(0, 3, 0)

	cert := New(domain, issuer, expiresAt)
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

func toDTO(c Cert) CertDto {
	today := time.Now()
	status := "valid"
	if c.ExpiresAt.Before(today) {
		status = "expired"
	} else if c.ExpiresAt.Before(today.AddDate(0, 0, 10)) {
		status = "expires_soon"
	}

	return CertDto{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt.Format(time.DateOnly),
		ExpiresAt: c.ExpiresAt.Format(time.DateOnly),
		Domain:    c.Domain.String(),
		Issuer:    c.Issuer.String(),
		Status:    status,
	}
}
