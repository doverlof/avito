package pull_request

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	repo2 "github.com/doverlof/avito_help/internal/client/repo"
	"github.com/doverlof/avito_help/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrPRExists         = errors.New("pull request already exists")
	ErrPRNotFound       = errors.New("pull request not found")
	ErrDontHaveReviewer = errors.New("dont have reviewers")
	ErrNoRowsAffected   = errors.New("no rows affected")
)

type Repo interface {
	Create(ctx context.Context, pullRequest model.CreatePullRequest, reviewers []model.User) error
	Merge(ctx context.Context, pullRequestID string) (model.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]model.PullRequest, error)
	GetByID(ctx context.Context, pullRequestID string) (model.PullRequest, error)
	ChangeReviewer(ctx context.Context, pullRequestID, oldReviewerID, reviewerID string) (model.PullRequest, error)
	GetUserStatistics(ctx context.Context) ([]model.UserStatistics, error)
}

type Selector interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type repo struct {
	sqlClient *sqlx.DB
}

func New(sqlClient *sqlx.DB) Repo {
	return &repo{
		sqlClient: sqlClient,
	}
}

func (r *repo) Create(ctx context.Context, pullRequest model.CreatePullRequest, reviewers []model.User) error {
	tx, err := r.sqlClient.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	query, args, err := sq.Insert("pull_requests").Columns(
		"pull_request_id",
		"pull_request_name",
		"author_id",
	).Values(
		pullRequest.PullRequestID,
		pullRequest.PullRequestName,
		pullRequest.AuthorID,
	).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return repo2.ErrToCreateToCreateSql(err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok && pqErr.Code == "23505" {
			// unique violation
			return ErrPRExists
		}
		return err
	}

	builder := sq.Insert("pr_reviewers").Columns("pull_request_id", "reviewer_id").PlaceholderFormat(sq.Dollar)
	for _, reviewer := range reviewers {
		builder = builder.Values(pullRequest.PullRequestID, reviewer.ID)
	}

	queryRev, argsRev, err := builder.ToSql()
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return ErrDontHaveReviewer
		}
		return err
	}

	_, err = tx.ExecContext(ctx, queryRev, argsRev...)
	if err != nil {
		return err
	}
	return nil
}

type pullRequestByReviewer struct {
	PullRequestID   string    `db:"pull_request_id"`
	PullRequestName string    `db:"pull_request_name"`
	AuthorID        string    `db:"author_id"`
	Status          string    `db:"status"`
	MergedAt        time.Time `db:"merged_at"`
	ReviewerID      string    `db:"reviewer_id"`
}

func (r *repo) Merge(ctx context.Context, pullRequestID string) (model.PullRequest, error) {
	query, args, err := sq.Update("pull_requests").Set("status", model.StatusMerge).
		Set("merged_at", time.Now()).
		Where(sq.Eq{"pull_request_id": pullRequestID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return model.PullRequest{}, repo2.ErrToCreateToCreateSql(err)
	}
	tx, err := r.sqlClient.Beginx()
	if err != nil {
		return model.PullRequest{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return model.PullRequest{}, err
	}
	v, err := res.RowsAffected()
	if err != nil {
		return model.PullRequest{}, err
	}
	if v == 0 {
		return model.PullRequest{}, ErrPRNotFound
	}

	return selectByID(ctx, tx, pullRequestID)
}

func convertPullRequest(rows []pullRequestByReviewer) model.PullRequest {
	reviewerIDs := make([]string, len(rows))
	for i, row := range rows {
		reviewerIDs[i] = row.ReviewerID
	}
	return model.PullRequest{
		AuthorID:        rows[0].AuthorID,
		PullRequestID:   rows[0].PullRequestID,
		PullRequestName: rows[0].PullRequestName,
		Status:          model.PullRequestStatus(rows[0].Status),
		MergedAt:        rows[0].MergedAt,
		ReviewerIDs:     reviewerIDs,
	}
}

func (r *repo) GetByReviewer(ctx context.Context, userID string) ([]model.PullRequest, error) {
	query, args, err := sq.Select(
		"pr.pull_request_id",
		"pr.pull_request_name",
		"pr.author_id",
		"pr.status",
	).From("pull_requests pr").
		Join("pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where(sq.Eq{"prr.reviewer_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, repo2.ErrToCreateToCreateSql(err)
	}

	var prs []model.PullRequest
	err = r.sqlClient.SelectContext(ctx, &prs, query, args...)
	if err != nil {
		return nil, err
	}

	if prs == nil {
		prs = []model.PullRequest{}
	}

	return prs, nil
}

func (r *repo) GetUserStatistics(ctx context.Context) ([]model.UserStatistics, error) {
	query := `
        SELECT 
            u.user_id,
            u.username,
            COALESCE(u.team_name, '') as team_name,
            u.is_active,
            
            COUNT(DISTINCT prr.pull_request_id) as total_review_assignments,
            COUNT(DISTINCT CASE WHEN pr_rev.status = $1 THEN prr.pull_request_id END) as open_review_assignments,
            COUNT(DISTINCT CASE WHEN pr_rev.status = $2 THEN prr.pull_request_id END) as merged_review_assignments,
            
            COUNT(DISTINCT pr_auth.pull_request_id) as total_authored_prs,
            COUNT(DISTINCT CASE WHEN pr_auth.status = $1 THEN pr_auth.pull_request_id END) as open_authored_prs,
            COUNT(DISTINCT CASE WHEN pr_auth.status = $2 THEN pr_auth.pull_request_id END) as merged_authored_prs
            
        FROM users u
        LEFT JOIN pr_reviewers prr ON u.user_id = prr.reviewer_id
        LEFT JOIN pull_requests pr_rev ON prr.pull_request_id = pr_rev.pull_request_id
        LEFT JOIN pull_requests pr_auth ON u.user_id = pr_auth.author_id
        GROUP BY u.user_id, u.username, u.team_name, u.is_active
        ORDER BY u.user_id
    `

	var stats []model.UserStatistics
	err := r.sqlClient.SelectContext(ctx, &stats, query, model.StatusOpen, model.StatusMerge)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	if stats == nil {
		stats = []model.UserStatistics{}
	}

	return stats, nil
}

func selectByID(ctx context.Context, selector Selector, pullRequestID string) (model.PullRequest, error) {
	query, args, err := sq.Select(
		"p.pull_request_id",
		"p.pull_request_name",
		"p.author_id",
		"p.status",
		"p.merged_at",
		"r.reviewer_id",
	).
		From("pull_requests p").
		LeftJoin("pr_reviewers r ON p.pull_request_id = r.pull_request_id").
		Where(sq.Eq{"p.pull_request_id": pullRequestID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return model.PullRequest{}, repo2.ErrToCreateToCreateSql(err)
	}
	var rows []pullRequestByReviewer
	err = selector.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return model.PullRequest{}, repo2.ErrToCreateToCreateSql(err)
	}
	if len(rows) == 0 {
		return model.PullRequest{}, ErrPRNotFound
	}

	return convertPullRequest(rows), nil
}

func (r *repo) GetByID(ctx context.Context, pullRequestID string) (model.PullRequest, error) {
	return selectByID(ctx, r.sqlClient, pullRequestID)
}

func (r *repo) ChangeReviewer(ctx context.Context, pullRequestID, oldReviewerID, reviewerID string) (model.PullRequest, error) {
	tx, err := r.sqlClient.Beginx()
	if err != nil {
		return model.PullRequest{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	//Update
	query, args, err := sq.Update("pr_reviewers").Set("reviewer_id", reviewerID).
		Where(sq.Eq{"pull_request_id": pullRequestID}, sq.Eq{"reviewer_id": oldReviewerID}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return model.PullRequest{}, repo2.ErrToCreateToCreateSql(err)
	}
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return model.PullRequest{}, err
	}
	v, err := res.RowsAffected()
	if err != nil {
		return model.PullRequest{}, err
	}
	if v == 0 {
		return model.PullRequest{}, ErrNoRowsAffected
	}
	return selectByID(ctx, tx, pullRequestID)
}
