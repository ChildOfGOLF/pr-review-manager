package service

import "pr-review-manager/internal/repository"

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}
