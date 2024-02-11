package certs

type Repo interface {
	Save(cert Cert) error
	GetAll() ([]Cert, error)
	Delete(id ID) error
}

type CertsRepo struct {
	db map[string]Cert
}

func NewRepo() *CertsRepo {
	return &CertsRepo{
		db: make(map[string]Cert),
	}
}

func (r *CertsRepo) Save(cert Cert) error {
	if r.isDuplicate(cert.Domain) {
		return ErrDuplicateDomain
	}

	r.db[cert.ID.String()] = cert
	return nil
}

func (r *CertsRepo) GetAll() ([]Cert, error) {
	certs := make([]Cert, 0, len(r.db))
	for _, cert := range r.db {
		certs = append(certs, cert)
	}
	return certs, nil
}

func (r *CertsRepo) Delete(id ID) error {
	delete(r.db, id.String())
	return nil
}

func (r *CertsRepo) isDuplicate(domain Domain) bool {
	for _, v := range r.db {
		if v.Domain.String() == domain.String() {
			return true
		}
	}
	return false
}
