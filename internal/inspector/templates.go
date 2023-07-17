package inspector

import (
	"bytes"
	"fmt"
	"html/template"
)

// ParseTemplate parses a template and returns the subject and body
func ParseTemplate(templateName string, data map[string]any) (string, string, error) {
	tmpl, err := template.New("email").ParseFiles(fmt.Sprintf("ui/html/emails/%s", templateName))
	if err != nil {
		return "", "", err
	}

	// Execute the "subject" template
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return "", "", err
	}

	// Execute the "body" template
	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return "", "", err
	}

	return subject.String(), body.String(), nil
}

// ParseFailedPingTemplate parses the corresponding template and returns the subject and body
func ParseFailedPingTemplate(fail FailedPing) (string, string, error) {
	return ParseTemplate("alert_ping.html.tmpl", map[string]any{
		"URL":      fail.URL,
		"Expected": fail.ExpectedCode,
		"Actual":   fail.ActualCode,
		"Time":     fail.Time.UTC().Format("2006-01-02 15:04:05"),
	})
}
