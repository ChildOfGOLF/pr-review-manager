package service

import (
	"context"

	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/errors"
	"pr-review-manager/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
	prRepo   *repository.PRRepository
}

func NewUserService(userRepo *repository.UserRepository, prRepo *repository.PRRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrNotFound
	}
	return user, nil
}

func (s *UserService) GetReview(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	prs, err := s.prRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}
	if prs == nil {
		prs = []domain.PullRequestShort{}
	}
	return prs, nil
}
