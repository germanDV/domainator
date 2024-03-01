package tlser

import (
	"crypto/tls"
	"net"
	"time"
)

type CertStatus string

const (
	StatusOK               CertStatus = "OK"
	StatusExpired          CertStatus = "Expired"
	StatusCannotConnect    CertStatus = "CannotConnect"
	StatusHostnameMismatch CertStatus = "HostnameMismatch"
	StatusIssuerNotFound   CertStatus = "IssuerNotFound"
)

type CertData struct {
	Status CertStatus
	Expiry time.Time
	Issuer string
}

type Client interface {
	GetCertData(domain string) CertData
}

type TLSer struct {
	timeout time.Duration
}

func New(timeout time.Duration) *TLSer {
	return &TLSer{timeout}
}

func (t TLSer) GetCertData(domain string) CertData {
	dialer := net.Dialer{Timeout: t.timeout}
	conn, err := tls.DialWithDialer(&dialer, "tcp", domain+":443", nil)
	if err != nil {
		return CertData{Status: StatusCannotConnect}
	}
	defer conn.Close()

	err = conn.VerifyHostname(domain)
	if err != nil {
		return CertData{Status: StatusHostnameMismatch}
	}

	now := time.Now()
	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	if expiry.Before(now) {
		return CertData{Status: StatusExpired, Expiry: expiry}
	}

	cas := conn.ConnectionState().PeerCertificates[0].Issuer.Organization
	if len(cas) == 0 {
		return CertData{Status: StatusIssuerNotFound}
	}

	return CertData{Status: StatusOK, Expiry: expiry, Issuer: cas[0]}
}
