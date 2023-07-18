package services

import "domainator/internal/notificators"

// NotificationPreference is a struct that represents a user's notification preference
type NotificationPreference struct {
	Service    notificators.Service
	Enabled    bool
	To         string // email address | slack channel
	WebhookURL string // slack webhook url
}
