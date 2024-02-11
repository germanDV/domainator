package certs

import "time"

type Service interface {
	RegisterCert(dto RegisterCertReq) (RegisterCertResp, error)
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
		Domain:    Domain("go.dev"),
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		Domain:    Domain("archlinux.org"),
	})
	repo.Save(Cert{
		ID:        NewID(),
		CreatedAt: time.Now(),
		Domain:    Domain("debian.org"),
	})

	return &CertsService{
		repo: repo,
	}
}

func (s *CertsService) RegisterCert(dto RegisterCertReq) (RegisterCertResp, error) {
	domain, err := NewDomain(dto.Domain)
	if err != nil {
		return RegisterCertResp{}, err
	}

	cert := New(domain)
	err = s.repo.Save(cert)
	if err != nil {
		return RegisterCertResp{}, err
	}

	return RegisterCertResp{ID: cert.ID.String()}, nil
}

func (s *CertsService) GetAll() ([]CertDto, error) {
	certificates, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	dtos := []CertDto{}
	for _, c := range certificates {
		dtos = append(dtos, CertDto{
			ID:        c.ID.String(),
			Domain:    c.Domain.String(),
			CreatedAt: c.CreatedAt,
		})
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
