package tlsermock

import (
	"strings"
	"time"

	"github.com/germandv/domainator/internal/tlser"
)

type MockTLSer struct{}

func New() *MockTLSer {
	return &MockTLSer{}
}

func (m MockTLSer) GetCertData(domain string) tlser.CertData {
	if strings.Contains(domain, "expired") {
		return tlser.CertData{
			Status: tlser.StatusExpired,
			Expiry: time.Now().Add(-24 * time.Hour),
			Issuer: "Test-Issuer",
		}
	}

	if strings.Contains(domain, "notconnect") {
		return tlser.CertData{
			Status: tlser.StatusCannotConnect,
			Expiry: time.Now().Add(6 * time.Hour),
			Issuer: "Test-Issuer",
		}
	}

	if strings.Contains(domain, "about") {
		return tlser.CertData{
			Status: tlser.StatusOK,
			Expiry: time.Now().Add(6 * time.Hour),
			Issuer: "Test-Issuer",
		}
	}

	return tlser.CertData{
		Status: tlser.StatusOK,
		Expiry: time.Now().Add(24 * 30 * time.Hour),
		Issuer: "Test-Issuer",
	}
}
