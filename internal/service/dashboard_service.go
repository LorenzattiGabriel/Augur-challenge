package service

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
)

type DashboardService struct {
	repo  *repository.DashboardRepository
	cache *cache.Cache
}

func NewDashboardService(repo *repository.DashboardRepository, c *cache.Cache) *DashboardService {
	return &DashboardService{
		repo:  repo,
		cache: c,
	}
}

func (s *DashboardService) GetSummary(ctx context.Context, timeRange string) (*model.DashboardSummary, error) {
	if timeRange == "" {
		timeRange = "7d"
	}
	validRanges := map[string]bool{"24h": true, "7d": true, "30d": true}
	if !validRanges[timeRange] {
		timeRange = "7d"
	}

	cacheKey := cache.GenerateKey("dashboard_summary", map[string]string{"range": timeRange})
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*model.DashboardSummary), nil
	}

	summary, err := s.repo.GetSummary(ctx, timeRange)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, summary, cache.TTLDashboardSummary)
	return summary, nil
}
