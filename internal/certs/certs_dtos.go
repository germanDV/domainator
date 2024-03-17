package certs

type RegisterCertReq struct {
	Domain string
	UserID string
}

type GetAllCertsReq struct {
	UserID string
}

type CertDto struct {
	ID        string
	CreatedAt string
	ExpiresAt string
	Domain    string
	Issuer    string
	Status    string
	Error     string
}
