package usecase

import (
	"context"
	"github.com/alexperezortuno/go-simple-auth/domain"
)

type UserRepository interface {
	Get(ctx context.Context, username string) (*domain.User, error)
	Save(ctx context.Context, user *domain.User) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo}
}

func (s *UserService) GetUser(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.Get(ctx, username)
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	return s.repo.Save(ctx, user)
}
