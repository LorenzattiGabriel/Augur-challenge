package service

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
)

type CampaignService struct {
	repo  repository.CampaignRepositoryInterface
	cache *cache.Cache
}

func NewCampaignService(repo repository.CampaignRepositoryInterface, c *cache.Cache) *CampaignService {
	return &CampaignService{
		repo:  repo,
		cache: c,
	}
}

func (s *CampaignService) GetIndicatorsTimeline(ctx context.Context, campaignID string, params model.TimelineParams) (*model.CampaignWithTimeline, error) {
	if params.GroupBy == "" {
		params.GroupBy = "day"
	}
	if params.GroupBy != "day" && params.GroupBy != "week" {
		params.GroupBy = "day"
	}

	cacheKey := cache.GenerateKey("campaign_timeline", map[string]interface{}{
		"id":         campaignID,
		"group_by":   params.GroupBy,
		"start_date": params.StartDate,
		"end_date":   params.EndDate,
	})
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*model.CampaignWithTimeline), nil
	}

	timeline, err := s.repo.GetIndicatorsTimeline(ctx, campaignID, params)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, timeline, cache.TTLCampaignTimeline)
	return timeline, nil
}
