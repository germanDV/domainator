package services

// NotificationPreference is a struct that represents a user's notification preference
type NotificationPreference struct {
	Service    string // email | slack
	Enabled    bool
	To         string // email address | slack channel
	WebhookURL string // slack webhook url
}
