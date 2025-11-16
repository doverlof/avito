package team

import (
	"context"
	"errors"

	teamRepo "github.com/doverlof/avito_help/internal/client/repo/team"
	"github.com/doverlof/avito_help/internal/model"
)

var (
	ErrTeamNotFound = errors.New("team not found or don't have members")
	ErrTeamExists   = errors.New("team already exists")
)

type UseCase interface {
	Add(ctx context.Context, team model.Team) error
	Get(ctx context.Context, name string) (model.Team, error)
}

type useCase struct {
	repo teamRepo.Repo
}

func New(repo teamRepo.Repo) UseCase {
	return &useCase{
		repo: repo,
	}
}

func (u *useCase) Add(ctx context.Context, model model.Team) error {
	err := u.repo.Add(ctx, model)
	if errors.Is(err, teamRepo.ErrTeamExists) {
		return ErrTeamExists
	}
	return err
}

func (u *useCase) Get(ctx context.Context, name string) (model.Team, error) {
	team, err := u.repo.Get(ctx, name)
	if errors.Is(err, teamRepo.ErrTeamNotFound) {
		return model.Team{}, ErrTeamNotFound
	}
	return team, err
}
