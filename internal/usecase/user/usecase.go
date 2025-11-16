package user

import (
	"context"
	"errors"

	userRepo "github.com/doverlof/avito_help/internal/client/repo/user"
	"github.com/doverlof/avito_help/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UseCase interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error)
	GetByID(ctx context.Context, userID string) (model.User, error)
}

type useCase struct {
	repo userRepo.Repo
}

func New(repo userRepo.Repo) UseCase {
	return &useCase{
		repo: repo,
	}
}

func (u *useCase) SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error) {
	user, err := u.repo.SetIsActive(ctx, userID, isActive)
	if errors.Is(err, userRepo.ErrUserNotFound) {
		return model.User{}, ErrUserNotFound
	}
	return user, err
}

func (u *useCase) GetByID(ctx context.Context, userID string) (model.User, error) {
	user, err := u.repo.GetByID(ctx, userID)
	if errors.Is(err, userRepo.ErrUserNotFound) {
		return model.User{}, ErrUserNotFound
	}
	return user, err
}
