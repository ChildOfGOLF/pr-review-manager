package repository

import (
	"context"
	"database/sql"
	"fmt"

	"pr-review-manager/internal/domain"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, tx *sql.Tx, teamName string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO teams (team_name) VALUES ($1)", teamName)
	return err
}

func (r *TeamRepository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	return exists, err
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("team not found")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1
		ORDER BY username
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (r *TeamRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
