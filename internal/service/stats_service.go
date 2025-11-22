package service

import (
	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
	}
}

func (s *StatsService) GetStats() (*domain.Stats, error) {
	return s.statsRepo.GetStats()
}
