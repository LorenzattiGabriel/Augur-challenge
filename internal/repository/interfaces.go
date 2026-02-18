package repository

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
)

type IndicatorRepositoryInterface interface {
	GetByID(ctx context.Context, id string) (*model.IndicatorWithRelations, error)
	Search(ctx context.Context, params model.SearchParams) (*model.SearchResult, error)
	GetIndicatorsByIDs(ctx context.Context, ids []string) ([]model.Indicator, error)
}

type CampaignRepositoryInterface interface {
	GetByID(ctx context.Context, id string) (*model.Campaign, error)
	GetIndicatorsTimeline(ctx context.Context, campaignID string, params model.TimelineParams) (*model.CampaignWithTimeline, error)
}

type DashboardRepositoryInterface interface {
	GetSummary(ctx context.Context, timeRange string) (*model.DashboardSummary, error)
}
