package notifier

import (
	"fmt"
	"time"
)

// SlackMessage is the payload required to send a slack message
type SlackMessage struct {
	WebHookURL string `json:"-"`
	UserName   string `json:"username,omitempty"`
	Channel    string `json:"channel,omitempty"`
	Text       string `json:"text,omitempty"`
}

// SlackNotifier is a Notifier that sends a slack message
type SlackNotifier struct {
	Username string
	Timeout  time.Duration
}

// NewSlacker returns a new SlackNotifier
func NewSlacker() Notifier[SlackMessage] {
	return &SlackNotifier{
		Username: "Domainator",
		Timeout:  5 * time.Second,
	}
}

// Notify sends a slack message
func (e *SlackNotifier) Notify(message SlackMessage) error {
	fmt.Println("@@@ TO BE IMPLEMENTED @@@")
	fmt.Println("Sending Slack message")
	fmt.Printf("\tChannel: %s\n", message.Channel)
	fmt.Printf("\tMessage: %+v\n", message.Text)
	fmt.Println()
	return nil
}
