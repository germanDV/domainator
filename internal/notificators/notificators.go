// Package notificators provides an enum with the supported notification services
package notificators

// Service is an enum that holds the supported options
type Service int

const (
	// Nil is the zero value for notificators.Service
	Nil Service = iota
	// Email is the email notification service
	Email
	// Slack is the slack notification service
	Slack
)

func (ns Service) String() string {
	switch ns {
	case Email:
		return "email"
	case Slack:
		return "slack"
	default:
		return ""
	}
}
