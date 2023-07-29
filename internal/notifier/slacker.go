package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackMessage is the payload required to send a slack message
type SlackMessage struct {
	WebHookURL string `json:"-"`
	Text       string `json:"text,omitempty"`
}

// SlackNotifier is a Notifier that sends a slack message
type SlackNotifier struct {
	Timeout time.Duration
}

// NewSlacker returns a new SlackNotifier
func NewSlacker() Notifier {
	return &SlackNotifier{
		Timeout: 5 * time.Second,
	}
}

// Notify sends a slack message
func (e *SlackNotifier) Notify(message Message) error {
	payload := SlackMessage{
		Text: message.Body,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, message.To, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

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
		return fmt.Errorf("Error (%d) sending slack msg: %s", resp.StatusCode, buf.String())
	}
	return nil
}
