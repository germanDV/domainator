package users

import (
	"context"
	"domainator/internal/config"
	"domainator/internal/notificators"
	"domainator/internal/validation"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is an interface that the users repository must implement.
type Repo interface {
	Create(ctx context.Context, user *User) (*User, string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetNotificationPrefsBySettings(ctx context.Context, settingsID uuid.UUID) ([]NotificationPref, error)
	GetNotificationPrefsByUserID(ctx context.Context, userID uuid.UUID) ([]NotificationPref, error)
	Verify(ctx context.Context, email string, code string) error
	CreateNotification(ctx context.Context, userID uuid.UUID, service notificators.Service, recipient string) (*NotificationPref, error)
	UpdateNotification(ctx context.Context, id int, userID uuid.UUID, recipient string) (*NotificationPref, error)
	ToggleNotification(ctx context.Context, id int, userID uuid.UUID) (bool, error)
}

// PostgresRepo is a repository that implements the Repo interface.
type PostgresRepo struct {
	DB *pgxpool.Pool
}

// NewPostgresRepo returns a new instance of a PostgresRepo.
func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		DB: db,
	}
}

// Create inserts the User in the database and generats a verification code
func (pg *PostgresRepo) Create(ctx context.Context, user *User) (*User, string, error) {
	q1 := `insert into users (id, email, password, created_at) values ($1, $2, $3, $4)`
	args1 := []any{user.ID, user.Email, user.Password.hash, user.CreatedAt}

	code := generateCode()
	hashedCode := hashCode(code)
	exp := config.GetDuration("VERIFICATION_CODE_EXP")
	q2 := `insert into verification_codes (user_id, email, code, expires_at) values ($1, $2, $3, $4)`
	args2 := []any{user.ID, user.Email, hashedCode, time.Now().Add(exp).UTC()}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, q1, args1...)
	if err != nil {
		tx.Rollback(ctx)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, "", validation.ErrDuplicateEmail
		}
		return nil, "", err
	}

	_, err = tx.Exec(ctx, q2, args2...)
	if err != nil {
		tx.Rollback(ctx)
		return nil, "", err
	}

	tx.Commit(ctx)
	return user, code, nil
}

// GetByID finds a user by ID
func (pg *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, email, password, activated, plan_id from users where id = $1`
	var user User
	err := pg.DB.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.PlanID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, validation.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail finds a user by email
func (pg *PostgresRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, email, password, activated, plan_id from users where email = $1`
	var user User
	err := pg.DB.QueryRow(ctx, q, strings.ToLower(email)).Scan(
		&user.ID,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.PlanID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, validation.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetNotificationPrefsBySettings returns a list of notification preferences for a given user, found by settings id
func (pg *PostgresRepo) GetNotificationPrefsBySettings(ctx context.Context, settingsID uuid.UUID) ([]NotificationPref, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select np.service, np.recipient
		from ping_settings ps
		join notification_preferences np
			on np.user_id = ps.user_id
		where ps.id = $1
			and np.enabled = true;
	`

	rows, err := pg.DB.Query(ctx, q, settingsID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := []NotificationPref{}

	for rows.Next() {
		var svc string
		var to string
		err = rows.Scan(&svc, &to)
		if err != nil {
			return nil, err
		}

		p := NotificationPref{}
		switch svc {
		case notificators.Email.String():
			p.Service = notificators.Email
		case notificators.Slack.String():
			p.Service = notificators.Slack
		default:
			p.Service = notificators.Nil
		}
		p.To = to
		p.Enabled = true

		prefs = append(prefs, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return prefs, nil
}

// GetNotificationPrefsByUserID returns a list of notification preferences for a given user, found by user id
func (pg *PostgresRepo) GetNotificationPrefsByUserID(ctx context.Context, userID uuid.UUID) ([]NotificationPref, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, service, recipient, enabled
		from notification_preferences
		where user_id = $1
		order by created_at desc;
	`

	rows, err := pg.DB.Query(ctx, q, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := []NotificationPref{}
	for rows.Next() {
		var pref NotificationPref
		var svc string
		err = rows.Scan(&pref.ID, &svc, &pref.To, &pref.Enabled)
		if err != nil {
			return nil, err
		}

		switch svc {
		case notificators.Email.String():
			pref.Service = notificators.Email
		case notificators.Slack.String():
			pref.Service = notificators.Slack
		default:
			pref.Service = notificators.Nil
		}

		prefs = append(prefs, pref)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return prefs, nil
}

func (pg *PostgresRepo) getVerificationCode(ctx context.Context, email string) (*VerificationCode, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	code := VerificationCode{
		Email: email,
	}

	q := `select code, expires_at from verification_codes where email = $1 order by created_at desc limit 1`
	err := pg.DB.QueryRow(ctx, q, strings.ToLower(email)).Scan(&code.Hash, &code.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &code, nil
}

// Verify checks the verification code provided by the user and marks the user as activated.
func (pg *PostgresRepo) Verify(ctx context.Context, email string, candidate string) error {
	code, err := pg.getVerificationCode(ctx, email)
	if err != nil {
		return validation.ErrInvalidCode
	}

	if time.Now().UTC().After(code.ExpiresAt) {
		return validation.ErrInvalidCode
	}

	match := code.Matches(candidate)
	if !match {
		return validation.ErrInvalidCode
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `update users set activated = true where email = $1`
	_, err = pg.DB.Exec(ctx, q, email)
	return err
}

// CreateNotification creates a new notification preference for a user and enables it
func (pg *PostgresRepo) CreateNotification(ctx context.Context, userID uuid.UUID, service notificators.Service, recipient string) (*NotificationPref, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into notification_preferences (user_id, service, enabled, recipient)
		values($1, $2, $3, $4)
		returning id, enabled, recipient;
	`
	args := []any{userID, service.String(), true, recipient}

	var pref NotificationPref
	err := pg.DB.QueryRow(ctx, q, args...).Scan(
		&pref.ID,
		&pref.Enabled,
		&pref.To,
	)
	if err != nil {
		return nil, err
	}

	pref.Service = service
	return &pref, nil
}

// UpdateNotification updates the recipient of the notification preference.
func (pg *PostgresRepo) UpdateNotification(ctx context.Context, id int, userID uuid.UUID, recipient string) (*NotificationPref, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `update notification_preferences
		set recipient = $1
		where id = $2 and user_id = $3
		returning id, enabled, recipient, service;
	`
	args := []any{recipient, id, userID}

	var svc string
	var pref NotificationPref
	err := pg.DB.QueryRow(ctx, q, args...).Scan(
		&pref.ID,
		&pref.Enabled,
		&pref.To,
		&svc,
	)
	if err != nil {
		return nil, err
	}

	switch svc {
	case notificators.Email.String():
		pref.Service = notificators.Email
	case notificators.Slack.String():
		pref.Service = notificators.Slack
	default:
		pref.Service = notificators.Nil
	}

	return &pref, nil
}

// ToggleNotification enables or disables a notification preference
func (pg *PostgresRepo) ToggleNotification(ctx context.Context, id int, userID uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `update notification_preferences
		set enabled = not enabled
		where id = $1 and user_id = $2
		returning enabled;
	`
	args := []any{id, userID}

	var enabled bool
	err := pg.DB.QueryRow(ctx, q, args...).Scan(&enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, validation.ErrNotFound
		}
		return false, err
	}

	return enabled, nil
}
