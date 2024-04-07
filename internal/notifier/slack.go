package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackNotifier struct {
	Timeout time.Duration
}

func NewSlacker() Notifier {
	return &SlackNotifier{
		Timeout: 5 * time.Second,
	}
}

type SlackMessage struct {
	WebHookURL string `json:"-"`
	Text       string `json:"text,omitempty"`
}

func (sn *SlackNotifier) Notify(to string, notification Notification) error {
	text := fmt.Sprintf(
		"*Domain: %s*\n*Status: %s*\n*Hours: %d*\n",
		notification.Domain,
		notification.Status,
		notification.Hours,
	)

	payload := SlackMessage{Text: text}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, to, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: sn.Timeout}
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
		return fmt.Errorf("error (%d) sending slack msg: %s", resp.StatusCode, buf.String())
	}

	return nil
}
