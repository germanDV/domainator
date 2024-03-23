package certs

import "time"

// repoCert represents a Cert in the Repository layer.
type repoCert struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Domain    string    `db:"domain"`
	Issuer    string    `db:"issuer"`
	Error     string    `db:"error"`
}
