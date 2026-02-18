package service

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockIndicatorRepository struct {
	mock.Mock
}

func (m *MockIndicatorRepository) GetByID(ctx context.Context, id string) (*model.IndicatorWithRelations, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.IndicatorWithRelations), args.Error(1)
}

func (m *MockIndicatorRepository) Search(ctx context.Context, params model.SearchParams) (*model.SearchResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SearchResult), args.Error(1)
}

func (m *MockIndicatorRepository) GetIndicatorsByIDs(ctx context.Context, ids []string) ([]model.Indicator, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Indicator), args.Error(1)
}

type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) GetByID(ctx context.Context, id string) (*model.Campaign, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) GetIndicatorsTimeline(ctx context.Context, campaignID string, params model.TimelineParams) (*model.CampaignWithTimeline, error) {
	args := m.Called(ctx, campaignID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CampaignWithTimeline), args.Error(1)
}

type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetSummary(ctx context.Context, timeRange string) (*model.DashboardSummary, error) {
	args := m.Called(ctx, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DashboardSummary), args.Error(1)
}
