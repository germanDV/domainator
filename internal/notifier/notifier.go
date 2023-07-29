// Package notifier provides an interface for sending notifications via various methods.
package notifier

// Message is a struct that contains the information needed to send a notification.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Notifier is an interface for sending notifications.
type Notifier interface {
	Notify(msg Message) error
}
