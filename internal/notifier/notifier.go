// Package notifier provides an interface for sending notifications via various methods
package notifier

// Notifier is an interface for sending notifications
type Notifier[T EmailMessage | SlackMessage] interface {
	Notify(message T) error
}
