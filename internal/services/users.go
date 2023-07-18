package services

import (
	"context"
	"domainator/internal/notificators"
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IUserService is an interface that the user service must implement
type IUserService interface {
	Validate(payload Validatable) bool
	New(email, password string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetNotificationPreferencesBySettings(ctx context.Context, settingsID uuid.UUID) ([]NotificationPreference, error)
}

// User is a struct that represents a user
type User struct {
	ID        uuid.UUID `form:"id"`
	Email     string    `form:"email"`
	Password  pwd       `form:"-"`
	Activated bool      `form:"activated"`
	CreatedAt time.Time `form:"created_at"`
}

// UserService is a service that implements the IUserService interface
type UserService struct {
	Validator *validator.Validate
	DB        *pgxpool.Pool
}

// Validate validates a struct
func (us *UserService) Validate(payload Validatable) bool {
	return payload.Validate(us.Validator)
}

// New returns a User struct, hashing the password.
func (us *UserService) New(email, password string) (*User, error) {
	hashedPwd, err := hashPwd(password, 12)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:        uuid.New(),
		Email:     email,
		Activated: false,
		CreatedAt: time.Now().UTC(),
		Password: pwd{
			plain: &password,
			hash:  hashedPwd,
		},
	}

	return user, nil
}

// Create inserts the User in the database.
func (us *UserService) Create(ctx context.Context, user *User) (*User, error) {
	q := `insert into users (id, email, password, created_at)
		values ($1, $2, $3, $4)`

	args := []any{user.ID, user.Email, user.Password.hash, user.CreatedAt}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := us.DB.Exec(ctx, q, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	return user, nil
}

// GetByID finds a user by ID
func (us *UserService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, email, password, activated from users where id = $1`
	var user User
	err := us.DB.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail finds a user by email
func (us *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, email, password, activated from users where email = $1`
	var user User
	err := us.DB.QueryRow(ctx, q, strings.ToLower(email)).Scan(
		&user.ID,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
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
		var svc string
		var to string
		var webhook string
		err = rows.Scan(&svc, &to, &webhook)
		if err != nil {
			return nil, err
		}

		p := NotificationPreference{}
		switch svc {
		case notificators.Email.String():
			p.Service = notificators.Email
		case notificators.Slack.String():
			p.Service = notificators.Slack
		default:
			p.Service = notificators.Nil
		}
		p.To = to
		p.WebhookURL = webhook
		p.Enabled = true

		prefs = append(prefs, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return prefs, nil
}
