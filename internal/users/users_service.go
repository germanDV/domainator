package users

import (
	"context"
)

type Service interface {
	Save(ctx context.Context, req SaveReq) (User, error)
	GetByEmail(ctx context.Context, req GetByEmailReq) (User, error)
	GetByID(ctx context.Context, req GetByIDReq) (User, error)
	SetWebhookURL(ctx context.Context, req SetWebhookReq) error
}

type UsersService struct {
	repo Repo
}

func NewService(repo Repo) *UsersService {
	return &UsersService{
		repo: repo,
	}
}

func (s *UsersService) Save(ctx context.Context, req SaveReq) (User, error) {
	user := New(req.Name, req.Email, req.IdentityProvider, req.IdentityProviderID)
	err := s.repo.Save(ctx, serviceToRepoAdapter(user))
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *UsersService) GetByEmail(ctx context.Context, req GetByEmailReq) (User, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return User{}, err
	}

	u, err := repoToServiceAdapter(user)
	if err != nil {
		return User{}, err
	}

	return u, nil
}

func (s *UsersService) GetByID(ctx context.Context, req GetByIDReq) (User, error) {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return User{}, err
	}

	u, err := repoToServiceAdapter(user)
	if err != nil {
		return User{}, err
	}

	return u, nil
}

func (s *UsersService) SetWebhookURL(ctx context.Context, req SetWebhookReq) error {
	return s.repo.SetWebhookURL(ctx, req.UserID, req.URL)
}
