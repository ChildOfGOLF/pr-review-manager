package service

import (
	"math/rand"

	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/errors"
	"pr-review-manager/internal/repository"
)

type PRService struct {
	prRepo   *repository.PRRepository
	userRepo *repository.UserRepository
}

func NewPRService(prRepo *repository.PRRepository, userRepo *repository.UserRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *PRService) CreatePR(prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.PRExists(prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrPRExists
	}

	author, err := s.userRepo.GetUser(authorID)
	if err != nil {
		return nil, err
	}
	if author == nil {
		return nil, errors.ErrNotFound
	}

	candidates, err := s.userRepo.GetActiveTeamMembers(author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	reviewers := selectRandomReviewers(candidates, 2)

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.StatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.prRepo.CreatePR(pr); err != nil {
		return nil, err
	}

	return s.prRepo.GetPRWithoutTx(prID)
}

func (s *PRService) MergePR(prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetPRWithoutTx(prID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, errors.ErrNotFound
	}

	if pr.Status == domain.StatusMerged {
		return pr, nil
	}

	return s.prRepo.MergePR(prID)
}

func (s *PRService) ReassignReviewer(prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetPRWithoutTx(prID)
	if err != nil {
		return nil, "", err
	}
	if pr == nil {
		return nil, "", errors.ErrNotFound
	}

	if pr.Status == domain.StatusMerged {
		return nil, "", errors.ErrPRMerged
	}

	isAssigned := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", errors.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetUser(oldReviewerID)
	if err != nil {
		return nil, "", err
	}
	if oldReviewer == nil {
		return nil, "", errors.ErrNotFound
	}

	candidates, err := s.userRepo.GetActiveTeamMembers(oldReviewer.TeamName, "")
	if err != nil {
		return nil, "", err
	}

	excludeIDs := append(pr.AssignedReviewers, pr.AuthorID)
	filteredCandidates := filterUsers(candidates, excludeIDs)

	if len(filteredCandidates) == 0 {
		return nil, "", errors.ErrNoCandidate
	}

	newReviewer := filteredCandidates[rand.Intn(len(filteredCandidates))]

	if err := s.prRepo.ReassignReviewer(prID, oldReviewerID, newReviewer.UserID); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.prRepo.GetPRWithoutTx(prID)
	return updatedPR, newReviewer.UserID, err
}

func selectRandomReviewers(candidates []domain.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	count := min(len(candidates), maxCount)
	
	shuffled := make([]domain.User, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].UserID
	}
	return reviewers
}

func filterUsers(users []domain.User, excludeIDs []string) []domain.User {
	excludeMap := make(map[string]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	filtered := []domain.User{}
	for _, user := range users {
		if !excludeMap[user.UserID] {
			filtered = append(filtered, user)
		}
	}
	return filtered
}
