package service

import "pr-review-manager/internal/repository"

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
