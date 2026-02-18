package service

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
)

type IndicatorServiceInterface interface {
	GetByID(ctx context.Context, id string) (*model.IndicatorWithRelations, error)
	Search(ctx context.Context, params model.SearchParams) (*model.SearchResult, error)
}

type CampaignServiceInterface interface {
	GetIndicatorsTimeline(ctx context.Context, campaignID string, params model.TimelineParams) (*model.CampaignWithTimeline, error)
}

type DashboardServiceInterface interface {
	GetSummary(ctx context.Context, timeRange string) (*model.DashboardSummary, error)
}
