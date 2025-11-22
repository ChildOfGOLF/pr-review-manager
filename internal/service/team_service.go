package service

import (
	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/errors"
	"pr-review-manager/internal/repository"
)

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

func (s *TeamService) AddTeam(team *domain.Team) (*domain.Team, error) {
	exists, err := s.teamRepo.TeamExists(team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrTeamExists
	}

	tx, err := s.teamRepo.BeginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.teamRepo.CreateTeam(tx, team.TeamName); err != nil {
		return nil, err
	}

	for _, member := range team.Members {
		user := &domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}
		if err := s.userRepo.UpsertUser(tx, user); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.teamRepo.GetTeam(team.TeamName)
}

func (s *TeamService) GetTeam(teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return team, nil
}
