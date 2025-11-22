package repository

import (
	"database/sql"

	"pr-review-manager/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpsertUser(tx *sql.Tx, user *domain.User) error {
	_, err := tx.Exec(`
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET username = $2, team_name = $3, is_active = $4
	`, user.UserID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *UserRepository) GetUser(userID string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(`
		UPDATE users 
		SET is_active = $2 
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`, userID, isActive).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetActiveTeamMembers(teamName, excludeUserID string) ([]domain.User, error) {
	rows, err := r.db.Query(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != $2
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) DeactivateTeamUsers(tx *sql.Tx, teamName string) ([]string, error) {
	rows, err := tx.Query(`
		UPDATE users 
		SET is_active = false 
		WHERE team_name = $1 AND is_active = true
		RETURNING user_id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func (r *UserRepository) GetActiveUsers(tx *sql.Tx) ([]domain.User, error) {
	rows, err := tx.Query(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE is_active = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
