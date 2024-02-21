package certs

type RegisterCertReq struct {
	Domain string
}

type CertDto struct {
	ID        string
	CreatedAt string
	ExpiresAt string
	Domain    string
	Issuer    string
	Status    string
}
