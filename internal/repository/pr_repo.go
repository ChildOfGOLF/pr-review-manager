package repository

import (
	"database/sql"
	"time"

	"pr-review-manager/internal/domain"
)

type PRRepository struct {
	db *sql.DB
}

func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}

func (r *PRRepository) GetPRsByReviewer(userID string) ([]domain.PullRequestShort, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		var createdAt time.Time
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

func (r *PRRepository) CreatePR(pr *domain.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	createdAt := time.Now()
	_, err = tx.Exec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, domain.StatusOpen, createdAt)
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(`
			INSERT INTO pr_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`, pr.PullRequestID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PRRepository) PRExists(prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).Scan(&exists)
	return exists, err
}

func (r *PRRepository) GetPR(prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	var createdAt, mergedAt sql.NullTime
	
	err := r.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	rows, err := r.db.Query(`
		SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, nil
}

func (r *PRRepository) MergePR(prID string) (*domain.PullRequest, error) {
	mergedAt := time.Now()
	_, err := r.db.Exec(`
		UPDATE pull_requests 
		SET status = $2, merged_at = $3
		WHERE pull_request_id = $1 AND status != $2
	`, prID, domain.StatusMerged, mergedAt)
	if err != nil {
		return nil, err
	}

	return r.GetPR(prID)
}

func (r *PRRepository) ReassignReviewer(prID, oldReviewerID, newReviewerID string) error {
	_, err := r.db.Exec(`
		UPDATE pr_reviewers 
		SET user_id = $3
		WHERE pull_request_id = $1 AND user_id = $2
	`, prID, oldReviewerID, newReviewerID)
	return err
}
