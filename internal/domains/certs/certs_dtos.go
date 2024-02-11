package certs

import "time"

type RegisterCertReq struct {
	Domain string
}

type RegisterCertResp struct {
	ID string
}

type CertDto struct {
	ID        string
	Domain    string
	CreatedAt time.Time
}
