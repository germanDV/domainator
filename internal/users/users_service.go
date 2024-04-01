package users

import "context"

type Service interface {
	Save(context.Context, SaveReq) (User, error)
	GetByEmail(context.Context, GetByEmailReq) (User, error)
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
