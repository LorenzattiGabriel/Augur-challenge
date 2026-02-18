package handler

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockIndicatorService struct {
	mock.Mock
}

func (m *MockIndicatorService) GetByID(ctx context.Context, id string) (*model.IndicatorWithRelations, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.IndicatorWithRelations), args.Error(1)
}

func (m *MockIndicatorService) Search(ctx context.Context, params model.SearchParams) (*model.SearchResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SearchResult), args.Error(1)
}

type MockCampaignService struct {
	mock.Mock
}

func (m *MockCampaignService) GetIndicatorsTimeline(ctx context.Context, campaignID string, params model.TimelineParams) (*model.CampaignWithTimeline, error) {
	args := m.Called(ctx, campaignID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CampaignWithTimeline), args.Error(1)
}

type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetSummary(ctx context.Context, timeRange string) (*model.DashboardSummary, error) {
	args := m.Called(ctx, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DashboardSummary), args.Error(1)
}
