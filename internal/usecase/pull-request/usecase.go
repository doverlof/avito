package pull_request

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"time"

	pullRequestPkg "github.com/doverlof/avito_help/internal/client/repo/pull-request"
	userPkg "github.com/doverlof/avito_help/internal/client/repo/user"
	"github.com/doverlof/avito_help/internal/model"
)

var (
	ErrPRExists             = errors.New("pull request already exists")
	ErrPRNotFound           = errors.New("pull request not found")
	ErrDontHaveReviewers    = errors.New("dont have reviewers")
	ErrPRAlreadyMerged      = errors.New("pull request already merged")
	ErrNotAssigned          = errors.New("reviewer is not assigned to this PR")
	ErrTeamOrAuthorNotFound = errors.New("team or author not found")
)

type UseCase interface {
	Create(ctx context.Context, pullRequest model.CreatePullRequest) error
	Merge(ctx context.Context, pullRequestID string) (model.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]model.PullRequest, error)
	Reassign(ctx context.Context, pullRequestID, oldReviewerID string) (model.PullRequest, string, error)
}

type useCase struct {
	pullRequestRepo pullRequestPkg.Repo
	userRepo        userPkg.Repo
}

func New(repo pullRequestPkg.Repo, userRepo userPkg.Repo) UseCase {
	return &useCase{
		pullRequestRepo: repo,
		userRepo:        userRepo,
	}
}

func (u *useCase) Create(ctx context.Context, pullRequest model.CreatePullRequest) error {
	allAvailable, err := u.userRepo.GetReviewersByAuthorID(ctx, pullRequest.AuthorID)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, userPkg.ErrTeamOrAuthorNotFound) {
			return ErrTeamOrAuthorNotFound
		}
		return err
	}
	if len(allAvailable) == 0 {
		return ErrTeamOrAuthorNotFound
	}
	reviewers := pickReviewer(allAvailable)

	//Create pr
	err = u.pullRequestRepo.Create(ctx, pullRequest, reviewers)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, pullRequestPkg.ErrPRExists) {
			return ErrPRExists
		}
		if errors.Is(err, pullRequestPkg.ErrDontHaveReviewer) {
			return ErrTeamOrAuthorNotFound
		}
		return err
	}
	return nil
}

func pickReviewer(all []model.User) []model.User {
	n := len(all)
	if n <= 2 {
		return all
	}

	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(n)

	return []model.User{all[perm[0]], all[perm[1]]}
}

func (u *useCase) Merge(ctx context.Context, pullRequestID string) (model.PullRequest, error) {
	pullRequest, err := u.pullRequestRepo.Merge(ctx, pullRequestID)
	if errors.Is(err, pullRequestPkg.ErrPRNotFound) {
		return model.PullRequest{}, ErrPRExists
	}
	return pullRequest, err
}

func (u *useCase) GetByReviewer(ctx context.Context, userID string) ([]model.PullRequest, error) {
	return u.pullRequestRepo.GetByReviewer(ctx, userID)
}

func (u *useCase) Reassign(ctx context.Context, pullRequestID, oldReviewerID string) (model.PullRequest, string, error) {
	pullRequest, err := u.pullRequestRepo.GetByID(ctx, pullRequestID)
	if err != nil {
		if errors.Is(err, pullRequestPkg.ErrPRNotFound) {
			return model.PullRequest{}, "", ErrPRNotFound
		}
		return model.PullRequest{}, "", err
	}
	if pullRequest.Status == model.StatusMerge {
		return model.PullRequest{}, "", ErrPRAlreadyMerged
	}
	if !slices.Contains(pullRequest.ReviewerIDs, oldReviewerID) {
		return model.PullRequest{}, "", ErrNotAssigned
	}
	allAvailable, err := u.userRepo.GetReviewersByAuthorID(ctx, pullRequest.AuthorID)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, userPkg.ErrTeamOrAuthorNotFound) {
			return model.PullRequest{}, "", ErrTeamOrAuthorNotFound
		}
		return model.PullRequest{}, "", err
	}
	if len(allAvailable) == 0 {
		return model.PullRequest{}, "", ErrDontHaveReviewers
	}

	validate := make([]model.User, 0, len(allAvailable))
	for _, reviewer := range allAvailable {
		if reviewer.ID != oldReviewerID {
			validate = append(validate, reviewer)
		}
	}
	if len(validate) == 0 {
		return model.PullRequest{}, "", ErrDontHaveReviewers
	}
	reviewers := pickReviewer(validate)
	pullRequest, err = u.pullRequestRepo.ChangeReviewer(ctx, pullRequestID, oldReviewerID, reviewers[0].ID)
	return pullRequest, reviewers[0].ID, err
}
