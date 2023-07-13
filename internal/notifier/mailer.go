package notifier

import (
	"bytes"
	"domainator/internal/config"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EmailMessage is the payload required to send an email
type EmailMessage struct {
	To      string `json:"to,omitempty"`
	Subject string `json:"subject,omitempty"`
	HTML    string `json:"html,omitempty"`
}

// emailPayload is the full payload required by the email provider's API
type emailPayload struct {
	From string `json:"from,omitempty"`
	EmailMessage
}

// EmailNotifier is a Notifier that sends emails
type EmailNotifier struct {
	APIKey   string
	Endpoint string
	From     string
	Timeout  time.Duration
}

// NewMailer creates a new EmailNotifier
func NewMailer() Notifier[EmailMessage] {
	return &EmailNotifier{
		APIKey:   config.GetString("RESEND_API_KEY"),
		Endpoint: "https://api.resend.com/emails",
		From:     "Domainator <onboarding@resend.dev>",
		Timeout:  5 * time.Second,
	}
}

// Notify sends an email
func (e *EmailNotifier) Notify(message EmailMessage) error {
	body, err := json.Marshal(emailPayload{
		From:         e.From,
		EmailMessage: message,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, e.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)

	client := &http.Client{Timeout: e.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error (%d) sending email: %s", resp.StatusCode, buf.String())
	}
	return nil
}
