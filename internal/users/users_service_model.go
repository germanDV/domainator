package users

import (
	"time"

	"github.com/germandv/domainator/internal/common"
)

type SaveReq struct {
	Email              Email
	Name               string
	IdentityProvider   string
	IdentityProviderID string
}

type GetByEmailReq struct {
	Email Email
}

type User struct {
	ID                 common.ID
	Email              Email
	Name               string
	IdentityProvider   string
	IdentityProviderID string
	CreatedAt          time.Time
}

func New(name string, email Email, identityProvider string, identityProviderID string) User {
	return User{
		ID:                 common.NewID(),
		Name:               name,
		Email:              email,
		IdentityProvider:   identityProvider,
		IdentityProviderID: identityProviderID,
		CreatedAt:          time.Now(),
	}
}

// serviceToRepoAdapter transforms a User from the Service layer to the Repository layer.
func serviceToRepoAdapter(user User) repoUser {
	return repoUser{
		ID:                 user.ID.String(),
		Name:               user.Name,
		Email:              user.Email.String(),
		IdentityProvider:   user.IdentityProvider,
		IdentityProviderID: user.IdentityProviderID,
		CreatedAt:          user.CreatedAt,
	}
}

// repoToServiceAdapter transforms a User from the Repository layer to the Service layer.
func repoToServiceAdapter(user repoUser) (User, error) {
	parsedID, err := common.ParseID(user.ID)
	if err != nil {
		return User{}, err
	}

	parsedEmail, err := ParseEmail(user.Email)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:                 parsedID,
		Name:               user.Name,
		Email:              parsedEmail,
		IdentityProvider:   user.IdentityProvider,
		IdentityProviderID: user.IdentityProviderID,
		CreatedAt:          user.CreatedAt,
	}, nil
}
