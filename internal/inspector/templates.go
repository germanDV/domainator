package inspector

import "domainator/internal/notifier"

// ParseFailedPingTemplate parses the corresponding template and returns the subject and body
func ParseFailedPingTemplate(fail FailedPing) (string, string, error) {
	return notifier.ParseTemplate("alert_ping.html.tmpl", map[string]any{
		"URL":      fail.URL,
		"Expected": fail.ExpectedCode,
		"Actual":   fail.ActualCode,
		"Time":     fail.Time.UTC().Format("2006-01-02 15:04:05"),
	})
}
