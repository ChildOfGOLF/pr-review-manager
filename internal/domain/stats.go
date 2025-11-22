package domain

type Stats struct {
	TotalPRs       int            `json:"total_prs"`
	OpenPRs        int            `json:"open_prs"`
	MergedPRs      int            `json:"merged_prs"`
	ReviewerStats  []ReviewerStat `json:"reviewer_stats"`
	PRStats        []PRStat       `json:"pr_stats"`
}

type ReviewerStat struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	TotalAssigned  int    `json:"total_assigned"`
	OpenAssigned   int    `json:"open_assigned"`
	MergedAssigned int    `json:"merged_assigned"`
}

type PRStat struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	Status          string `json:"status"`
	ReviewersCount  int    `json:"reviewers_count"`
}
