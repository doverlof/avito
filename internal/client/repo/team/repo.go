package team

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	repo2 "github.com/doverlof/avito_help/internal/client/repo"
	"github.com/doverlof/avito_help/internal/convert"
	"github.com/doverlof/avito_help/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repo interface {
	Add(ctx context.Context, team model.Team) error
	Get(ctx context.Context, name string) (model.Team, error)
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
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found or don't have members")
	updateUsers     = `
        WITH data AS (
            SELECT *
            FROM unnest(
                $1::text[],
                $2::text[],
                $3::boolean[],
                $4::text[]
            ) AS t(user_id, username, is_active, team_name)
        )
        UPDATE users AS u
        SET 
            username  = d.username,
            is_active = d.is_active,
            team_name = d.team_name
        FROM data d
        WHERE u.user_id = d.user_id;
    `
)

func (r *repo) Add(ctx context.Context, team model.Team) error {
	query, args, err := sq.Insert("teams").Columns(
		"team_name",
	).Values(
		team.Name,
	).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return repo2.ErrToCreateToCreateSql(err)
	}
	userIDs := make([]string, len(team.Members))
	usernames := make([]string, len(team.Members))
	isActives := make([]bool, len(team.Members))
	teamNames := make([]string, len(team.Members))

	for i, member := range team.Members {
		userIDs[i] = member.ID
		usernames[i] = member.Name
		isActives[i] = member.IsActive
		teamNames[i] = team.Name
	}
	tx, err := r.sqlClient.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok && pqErr.Code == "23505" {
			// unique violation
			return ErrTeamExists
		}
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		updateUsers,
		pq.Array(userIDs),
		pq.Array(usernames),
		pq.Array(isActives),
		pq.Array(teamNames),
	)
	if err != nil {
		return err
	}
	return nil
}

type user struct {
	ID       string `db:"user_id"`
	Name     string `db:"username"`
	IsActive bool   `db:"is_active"`
}

func (r *repo) Get(ctx context.Context, name string) (model.Team, error) {
	query, args, err := sq.Select("user_id", "username", "is_active").From("users").
		Where(sq.Eq{"team_name": name}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return model.Team{}, repo2.ErrToCreateToCreateSql(err)
	}
	var users []user
	err = r.sqlClient.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return model.Team{}, fmt.Errorf("failed to query users: %w", err)
	}

	if len(users) == 0 {
		return model.Team{}, ErrTeamNotFound
	}
	return model.Team{
		Name:    name,
		Members: convert.Many(convertUsers, users),
	}, nil
}

func convertUsers(user user) model.Member {
	return model.Member{
		ID:       user.ID,
		Name:     user.Name,
		IsActive: user.IsActive,
	}
}
