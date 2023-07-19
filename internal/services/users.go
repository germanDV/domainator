package services

import (
	"context"
	"domainator/internal/config"
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
	Create(ctx context.Context, user *User) (*User, string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetNotificationPreferencesBySettings(ctx context.Context, settingsID uuid.UUID) ([]NotificationPreference, error)
	Verify(ctx context.Context, email string, code string) error
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

// Create inserts the User in the database and generats a verification code
func (us *UserService) Create(ctx context.Context, user *User) (*User, string, error) {
	q1 := `insert into users (id, email, password, created_at) values ($1, $2, $3, $4)`
	args1 := []any{user.ID, user.Email, user.Password.hash, user.CreatedAt}

	code := generateCode()
	hashedCode := hashCode(code)
	exp := config.GetDuration("VERIFICATION_CODE_EXP")
	q2 := `insert into verification_codes (user_id, email, code, expires_at) values ($1, $2, $3, $4)`
	args2 := []any{user.ID, user.Email, hashedCode, time.Now().Add(exp).UTC()}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := us.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, q1, args1...)
	if err != nil {
		tx.Rollback(ctx)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, "", ErrDuplicateEmail
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

func (us *UserService) getVerificationCode(ctx context.Context, email string) (*VerificationCode, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	code := VerificationCode{
		Email: email,
	}

	q := `select code, expires_at from verification_codes where email = $1 order by created_at desc limit 1`
	err := us.DB.QueryRow(ctx, q, strings.ToLower(email)).Scan(&code.Hash, &code.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &code, nil
}

// Verify checks the verification code provided by the user and marks the user as activated
func (us *UserService) Verify(ctx context.Context, email string, candidate string) error {
	code, err := us.getVerificationCode(ctx, email)
	if err != nil {
		return ErrInvalidCode
	}

	if time.Now().UTC().After(code.ExpiresAt) {
		return ErrInvalidCode
	}

	match := code.Matches(candidate)
	if !match {
		return ErrInvalidCode
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `update users set activated = true where email = $1`
	_, err = us.DB.Exec(ctx, q, email)
	return err
}
