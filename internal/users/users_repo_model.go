package users

import "time"

// repoUser represents a User in the Repository layer.
type repoUser struct {
	ID                 string    `db:"id"`
	Name               string    `db:"name"`
	Email              string    `db:"email"`
	Avatar             string    `db:"avatar_url"`
	CreatedAt          time.Time `db:"created_at"`
	IdentityProvider   string    `db:"identity_provider"`
	IdentityProviderID string    `db:"identity_provider_id"`
	WebhookURL         string    `db:"webhook_url"`
}
