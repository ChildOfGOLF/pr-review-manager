package service

import (
	"context"

	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/errors"
	"pr-review-manager/internal/repository"
)

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
	prRepo   *repository.PRRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository, prRepo *repository.PRRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *TeamService) AddTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	exists, err := s.teamRepo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrTeamExists
	}

	tx, err := s.teamRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.teamRepo.CreateTeam(ctx, tx, team.TeamName); err != nil {
		return nil, err
	}

	for _, member := range team.Members {
		user := &domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}
		if err := s.userRepo.UpsertUser(ctx, tx, user); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.teamRepo.GetTeam(ctx, team.TeamName)
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return team, nil
}

// DeactivateTeam массово деактивирует команду и безопасно переназначает открытые PR
// Все операции выполняются атомарно в одной транзакции
func (s *TeamService) DeactivateTeam(ctx context.Context, teamName string) (int, int, error) {
	tx, err := s.teamRepo.BeginTx(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	deactivatedUserIDs, err := s.userRepo.DeactivateTeamUsers(ctx, tx, teamName)
	if err != nil {
		return 0, 0, err
	}

	affectedPRs := 0
	if len(deactivatedUserIDs) > 0 {
		prIDs, err := s.prRepo.GetOpenPRsWithDeactivatedReviewers(ctx, tx, deactivatedUserIDs)
		if err != nil {
			return 0, 0, err
		}

		if len(prIDs) == 0 {
			if err := tx.Commit(); err != nil {
				return 0, 0, err
			}
			return len(deactivatedUserIDs), 0, nil
		}

		// Batch-удаление деактивированных ревьюверов
		if err := s.prRepo.RemoveDeactivatedReviewersFromAllPRs(ctx, tx, deactivatedUserIDs); err != nil {
			return 0, 0, err
		}

		prsMap, err := s.prRepo.GetPRsWithReviewers(ctx, tx, prIDs)
		if err != nil {
			return 0, 0, err
		}

		allActiveUsers, err := s.userRepo.GetActiveUsers(ctx, tx)
		if err != nil {
			return 0, 0, err
		}

		deactivatedMap := make(map[string]bool)
		for _, id := range deactivatedUserIDs {
			deactivatedMap[id] = true
		}

		// Сборка назначения для batch-вставки
		type Assignment struct{ PRID, UserID string }
		assignments := []Assignment{}

		for prID, pr := range prsMap {
			if pr.Status != "OPEN" {
				continue
			}

			currentReviewersCount := len(pr.AssignedReviewers)
			needed := 2 - currentReviewersCount

			if needed > 0 {
				excludeIDs := make(map[string]bool)
				excludeIDs[pr.AuthorID] = true
				for _, r := range pr.AssignedReviewers {
					excludeIDs[r] = true
				}

				candidates := []domain.User{}
				for _, user := range allActiveUsers {
					if !excludeIDs[user.UserID] {
						candidates = append(candidates, user)
					}
				}

				if len(candidates) > 0 {
					newReviewers := selectRandomReviewers(candidates, needed)
					for _, reviewerID := range newReviewers {
						assignments = append(assignments, Assignment{PRID: prID, UserID: reviewerID})
					}
				}
			}

			affectedPRs++
		}

		// Batch-вставка новых ревьюверов
		if len(assignments) > 0 {
			batchAssignments := make([]struct{ PRID, UserID string }, len(assignments))
			for i, a := range assignments {
				batchAssignments[i] = struct{ PRID, UserID string }{PRID: a.PRID, UserID: a.UserID}
			}
			if err := s.prRepo.BatchAddReviewers(ctx, tx, batchAssignments); err != nil {
				return 0, 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return len(deactivatedUserIDs), affectedPRs, nil
}
