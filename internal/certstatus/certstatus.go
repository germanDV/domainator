// Package certstatus provides an enum with the possible statuses of a TLS certificate.
package certstatus

// Status is an enum that holds the supported options.
type Status int

const (
	// Nil is the zero value for certstatus.Status.
	Nil Status = iota
	// OK is the certificate status when everything is fine.
	OK
	// AboutToExpire is the certificate status when it's about to expire.
	AboutToExpire
	// Expired is the certificate status when it's already expired.
	Expired
	// CannotConnect is the certificate status when the TLS connection cannot be established.
	CannotConnect
	// HostnameMismatch is the certificate status when the hostname in the certificate doesn't match the domain.
	HostnameMismatch
)

func (s Status) String() string {
	switch s {
	case OK:
		return "OK"
	case AboutToExpire:
		return "AboutToExpire"
	case Expired:
		return "Expired"
	case CannotConnect:
		return "CannotConnect"
	case HostnameMismatch:
		return "HostnameMismatch"
	default:
		return ""
	}
}
