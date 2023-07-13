package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationPreference is a struct that represents a user's notification preference
type NotificationPreference struct {
	Service    string // email | slack
	Enabled    bool
	To         string // email address | slack channel
	WebhookURL string // slack webhook url
}

// IUserService is an interface that the user service must implement
type IUserService interface {
	GetNotificationPreferencesBySettings(ctx context.Context, settingsID uuid.UUID) ([]NotificationPreference, error)
}

// UserService is a service that implements the IUserService interface
type UserService struct {
	DB *pgxpool.Pool
}

// GetNotificationPreferencesBySettings returns a list of notification preferences for a given user, found by settings id
func (us *UserService) GetNotificationPreferencesBySettings(ctx context.Context, settingsID uuid.UUID) ([]NotificationPreference, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select
			np.service,
			np.recipient,
			coalesce(np.webhook_url, '') as webhook_url
		from ping_settings ps
		join notification_preferences np
			on np.user_id = ps.user_id
		where ps.id = $1
			and np.enabled = true;
	`

	rows, err := us.DB.Query(ctx, q, settingsID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := []NotificationPreference{}

	for rows.Next() {
		p := NotificationPreference{}
		err = rows.Scan(&p.Service, &p.To, &p.WebhookURL)
		if err != nil {
			return nil, err
		}
		p.Enabled = true
		prefs = append(prefs, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return prefs, nil
}
