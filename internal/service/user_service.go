package service

import "pr-review-manager/internal/repository"

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
