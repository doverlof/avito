package stats

import (
	"context"

	"github.com/doverlof/avito_help/internal/client/repo/pull-request"
	"github.com/doverlof/avito_help/internal/model"
)

type UseCase interface {
	GetUserStatistics(ctx context.Context) ([]model.UserStatistics, error)
}

type useCase struct {
	prRepo pull_request.Repo
}

func New(prRepo pull_request.Repo) UseCase {
	return &useCase{
		prRepo: prRepo,
	}
}

func (u *useCase) GetUserStatistics(ctx context.Context) ([]model.UserStatistics, error) {
	return u.prRepo.GetUserStatistics(ctx)
}
