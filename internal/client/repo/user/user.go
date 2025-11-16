package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	repo2 "github.com/doverlof/avito_help/internal/client/repo"
	"github.com/doverlof/avito_help/internal/convert"
	"github.com/doverlof/avito_help/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repo interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error)
	GetByID(ctx context.Context, userID string) (model.User, error)
	GetReviewersByAuthorID(ctx context.Context, authorID string) ([]model.User, error)
}

type repo struct {
	sqlClient *sqlx.DB
}

func New(sqlClient *sqlx.DB) Repo {
	return &repo{
		sqlClient: sqlClient,
	}
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrTeamOrAuthorNotFound = errors.New("team or author not found")
)

type userDB struct {
	ID       string `db:"user_id"`
	Name     string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

func (r *repo) SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error) {
	query, args, err := sq.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return model.User{}, repo2.ErrToCreateToCreateSql(err)
	}

	result, err := r.sqlClient.ExecContext(ctx, query, args...)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return model.User{}, ErrUserNotFound
	}

	return r.GetByID(ctx, userID)
}

func (r *repo) GetByID(ctx context.Context, userID string) (model.User, error) {
	query, args, err := sq.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return model.User{}, repo2.ErrToCreateToCreateSql(err)
	}

	var user userDB
	err = r.sqlClient.GetContext(ctx, &user, query, args...)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return convertUser(user), nil
}

func convertUser(u userDB) model.User {
	return model.User{
		ID:       u.ID,
		Name:     u.Name,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func (r *repo) GetReviewersByAuthorID(ctx context.Context, authorID string) ([]model.User, error) {
	subQuery := sq.Select("team_name").From("users").
		Where(sq.Eq{"user_id": authorID})
	query, args, err := sq.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Expr("team_name = (?)", subQuery), sq.NotEq{"user_id": authorID}, sq.Eq{"is_active": true}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return []model.User{}, repo2.ErrToCreateToCreateSql(err)
	}

	var users []userDB
	err = r.sqlClient.SelectContext(ctx, &users, query, args...)
	if err != nil {
		fmt.Println(err)
		return []model.User{}, ErrTeamOrAuthorNotFound
	}
	return convert.Many(convertUser, users), nil
}
