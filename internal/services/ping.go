package services

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
)

// TODO: use proper UUIDv4
type uuid = string

// Pinger is an interface that the ping service must implement
type Pinger interface {
	SaveSettings(userID uuid, domain string, successCode int) (uuid, error)
	GetSummary(userID uuid) ([]*PingSummary, error)
	InsertPing(settingsID uuid, status int, took time.Duration) (uuid, error)
	Validate(payload Validatable) bool
}

// PingSummary returns data about a ping settings and the latest check
type PingSummary struct {
	ID        uuid
	Domain    string
	Status    string // 'healthy' or 'unhealthy'
	LastCheck time.Time
}

// PingService is a ping service
type PingService struct {
	Validator *validator.Validate
}

// Validate validates a struct
func (ps *PingService) Validate(payload Validatable) bool {
	return payload.Validate(ps.Validator)
}

// GetSummary returns all ping settings for a user with data about the latest ping
func (ps *PingService) GetSummary(userID uuid) ([]*PingSummary, error) {
	pings := []*PingSummary{
		{ID: "1", Domain: "debian.org", Status: "healthy", LastCheck: time.Now()},
		{ID: "2", Domain: "wikipedia.com", Status: "unhealthy", LastCheck: time.Now()},
		{ID: "3", Domain: "duckduckgo.com", Status: "healthy", LastCheck: time.Now()},
	}
	return pings, nil
}

// SaveSettings saves the settings for a domain to be pinged
func (ps *PingService) SaveSettings(userID uuid, domain string, successCode int) (uuid, error) {
	return "", errors.New("not implemented")
}

// InsertPing saves the results of an individual ping to a domain
func (ps *PingService) InsertPing(settingsID uuid, status int, took time.Duration) (uuid, error) {
	return "", errors.New("not implemented")
}
