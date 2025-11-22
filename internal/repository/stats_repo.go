package repository

import (
	"database/sql"

	"pr-review-manager/internal/domain"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetStats() (*domain.Stats, error) {
	stats := &domain.Stats{}

	err := r.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'OPEN' THEN 1 ELSE 0 END) as open,
			SUM(CASE WHEN status = 'MERGED' THEN 1 ELSE 0 END) as merged
		FROM pull_requests
	`).Scan(&stats.TotalPRs, &stats.OpenPRs, &stats.MergedPRs)
	if err != nil {
		return nil, err
	}

	reviewerStats, err := r.getReviewerStats()
	if err != nil {
		return nil, err
	}
	stats.ReviewerStats = reviewerStats

	prStats, err := r.getPRStats()
	if err != nil {
		return nil, err
	}
	stats.PRStats = prStats

	return stats, nil
}

func (r *StatsRepository) getReviewerStats() ([]domain.ReviewerStat, error) {
	rows, err := r.db.Query(`
		SELECT 
			u.user_id,
			u.username,
			COUNT(prr.pull_request_id) as total_assigned,
			SUM(CASE WHEN pr.status = 'OPEN' THEN 1 ELSE 0 END) as open_assigned,
			SUM(CASE WHEN pr.status = 'MERGED' THEN 1 ELSE 0 END) as merged_assigned
		FROM users u
		LEFT JOIN pr_reviewers prr ON u.user_id = prr.user_id
		LEFT JOIN pull_requests pr ON prr.pull_request_id = pr.pull_request_id
		GROUP BY u.user_id, u.username
		HAVING COUNT(prr.pull_request_id) > 0
		ORDER BY total_assigned DESC, u.username
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.ReviewerStat
	for rows.Next() {
		var stat domain.ReviewerStat
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.TotalAssigned, &stat.OpenAssigned, &stat.MergedAssigned); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if stats == nil {
		stats = []domain.ReviewerStat{}
	}

	return stats, nil
}

func (r *StatsRepository) getPRStats() ([]domain.PRStat, error) {
	rows, err := r.db.Query(`
		SELECT 
			pr.pull_request_id,
			pr.pull_request_name,
			pr.status,
			COUNT(prr.user_id) as reviewers_count
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		GROUP BY pr.pull_request_id, pr.pull_request_name, pr.status
		ORDER BY pr.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.PRStat
	for rows.Next() {
		var stat domain.PRStat
		if err := rows.Scan(&stat.PullRequestID, &stat.PullRequestName, &stat.Status, &stat.ReviewersCount); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if stats == nil {
		stats = []domain.PRStat{}
	}

	return stats, nil
}
