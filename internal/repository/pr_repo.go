package repository

import (
	"database/sql"
	"fmt"
	"strings"
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

func (r *PRRepository) GetOpenPRsWithDeactivatedReviewers(tx *sql.Tx, deactivatedUserIDs []string) ([]string, error) {
	if len(deactivatedUserIDs) == 0 {
		return nil, nil
	}
	
	placeholders := make([]string, len(deactivatedUserIDs))
	args := make([]interface{}, len(deactivatedUserIDs)+1)
	args[0] = "OPEN"
	for i, userID := range deactivatedUserIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = userID
	}
	
	query := fmt.Sprintf(`
		SELECT DISTINCT pr.pull_request_id
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE pr.status = $1 AND prr.user_id IN (%s)
	`, strings.Join(placeholders, ","))
	
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var prIDs []string
	for rows.Next() {
		var prID string
		if err := rows.Scan(&prID); err != nil {
			return nil, err
		}
		prIDs = append(prIDs, prID)
	}
	return prIDs, nil
}

func (r *PRRepository) RemoveReviewers(tx *sql.Tx, prID string, reviewerIDs []string) error {
	if len(reviewerIDs) == 0 {
		return nil
	}
	
	placeholders := make([]string, len(reviewerIDs))
	args := make([]interface{}, len(reviewerIDs)+1)
	args[0] = prID
	for i, reviewerID := range reviewerIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = reviewerID
	}
	
	query := fmt.Sprintf(`
		DELETE FROM pr_reviewers 
		WHERE pull_request_id = $1 AND user_id IN (%s)
	`, strings.Join(placeholders, ","))
	
	_, err := tx.Exec(query, args...)
	return err
}

func (r *PRRepository) RemoveDeactivatedReviewersFromAllPRs(tx *sql.Tx, deactivatedUserIDs []string) error {
	if len(deactivatedUserIDs) == 0 {
		return nil
	}
	
	placeholders := make([]string, len(deactivatedUserIDs))
	args := make([]interface{}, len(deactivatedUserIDs))
	for i, userID := range deactivatedUserIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = userID
	}
	
	query := fmt.Sprintf(`
		DELETE FROM pr_reviewers 
		WHERE user_id IN (%s)
		AND pull_request_id IN (
			SELECT pull_request_id FROM pull_requests WHERE status = 'OPEN'
		)
	`, strings.Join(placeholders, ","))
	
	_, err := tx.Exec(query, args...)
	return err
}

func (r *PRRepository) GetPRReviewers(tx *sql.Tx, prID string) ([]string, error) {
	rows, err := tx.Query(`
		SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}
	return reviewers, nil
}

func (r *PRRepository) GetPR(tx *sql.Tx, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	var createdAt time.Time
	var mergedAt sql.NullTime

	err := tx.QueryRow(`
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

	pr.CreatedAt = &createdAt
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	reviewers, err := r.GetPRReviewers(tx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
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

func (r *PRRepository) GetPRWithoutTx(prID string) (*domain.PullRequest, error) {
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

	return r.GetPRWithoutTx(prID)
}

func (r *PRRepository) ReassignReviewer(prID, oldReviewerID, newReviewerID string) error {
	_, err := r.db.Exec(`
		UPDATE pr_reviewers 
		SET user_id = $3
		WHERE pull_request_id = $1 AND user_id = $2
	`, prID, oldReviewerID, newReviewerID)
	return err
}

func (r *PRRepository) AddReviewer(tx *sql.Tx, prID, userID string) error {
	_, err := tx.Exec(`
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
	`, prID, userID)
	return err
}

func (r *PRRepository) BatchAddReviewers(tx *sql.Tx, assignments []struct{ PRID, UserID string }) error {
	if len(assignments) == 0 {
		return nil
	}
	
	valueStrings := make([]string, len(assignments))
	valueArgs := make([]interface{}, len(assignments)*2)
	
	for i, assignment := range assignments {
		valueStrings[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		valueArgs[i*2] = assignment.PRID
		valueArgs[i*2+1] = assignment.UserID
	}
	
	query := fmt.Sprintf(`
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES %s
	`, strings.Join(valueStrings, ","))
	
	_, err := tx.Exec(query, valueArgs...)
	return err
}

func (r *PRRepository) GetPRsWithReviewers(tx *sql.Tx, prIDs []string) (map[string]*domain.PullRequest, error) {
	if len(prIDs) == 0 {
		return make(map[string]*domain.PullRequest), nil
	}
	
	placeholders := make([]string, len(prIDs))
	args := make([]interface{}, len(prIDs))
	for i, prID := range prIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = prID
	}
	
	query := fmt.Sprintf(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id IN (%s)
	`, strings.Join(placeholders, ","))
	
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	prs := make(map[string]*domain.PullRequest)
	for rows.Next() {
		var pr domain.PullRequest
		var createdAt time.Time
		var mergedAt sql.NullTime
		
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt); err != nil {
			return nil, err
		}
		
		pr.CreatedAt = &createdAt
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}
		pr.AssignedReviewers = []string{}
		prs[pr.PullRequestID] = &pr
	}
	
	reviewerQuery := fmt.Sprintf(`
		SELECT pull_request_id, user_id
		FROM pr_reviewers
		WHERE pull_request_id IN (%s)
	`, strings.Join(placeholders, ","))
	
	reviewerRows, err := tx.Query(reviewerQuery, args...)
	if err != nil {
		return nil, err
	}
	defer reviewerRows.Close()
	
	for reviewerRows.Next() {
		var prID, userID string
		if err := reviewerRows.Scan(&prID, &userID); err != nil {
			return nil, err
		}
		if pr, ok := prs[prID]; ok {
			pr.AssignedReviewers = append(pr.AssignedReviewers, userID)
		}
	}
	
	return prs, nil
}
