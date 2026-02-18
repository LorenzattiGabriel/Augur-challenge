package service

import (
	"context"
	"testing"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCampaignService(t *testing.T) (*CampaignService, *MockCampaignRepository, *cache.Cache) {
	mockRepo := new(MockCampaignRepository)
	c, err := cache.New(cache.Config{MaxSizeMB: 10})
	require.NoError(t, err)
	svc := NewCampaignService(mockRepo, c)
	return svc, mockRepo, c
}

func TestCampaignService_GetIndicatorsTimeline_Success(t *testing.T) {
	svc, mockRepo, _ := setupCampaignService(t)
	ctx := context.Background()

	params := model.TimelineParams{GroupBy: "day"}
	expected := &model.CampaignWithTimeline{
		Campaign: model.CampaignDetail{
			ID:     "camp-1",
			Name:   "Operation Test",
			Status: "active",
		},
		Timeline: []model.TimelinePeriod{
			{Period: "2024-01-01", Counts: map[string]int{"ip": 5}},
		},
		Summary: model.TimelineSummary{TotalIndicators: 5},
	}

	mockRepo.On("GetIndicatorsTimeline", ctx, "camp-1", params).Return(expected, nil)

	result, err := svc.GetIndicatorsTimeline(ctx, "camp-1", params)

	assert.NoError(t, err)
	assert.Equal(t, "camp-1", result.Campaign.ID)
	assert.Equal(t, "Operation Test", result.Campaign.Name)
	assert.Len(t, result.Timeline, 1)
	mockRepo.AssertExpectations(t)
}

func TestCampaignService_GetIndicatorsTimeline_DefaultsGroupBy(t *testing.T) {
	svc, mockRepo, _ := setupCampaignService(t)
	ctx := context.Background()

	inputParams := model.TimelineParams{GroupBy: ""}
	expectedParams := model.TimelineParams{GroupBy: "day"}

	mockRepo.On("GetIndicatorsTimeline", ctx, "camp-1", expectedParams).Return(&model.CampaignWithTimeline{}, nil)

	_, err := svc.GetIndicatorsTimeline(ctx, "camp-1", inputParams)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCampaignService_GetIndicatorsTimeline_InvalidGroupByDefaultsToDay(t *testing.T) {
	svc, mockRepo, _ := setupCampaignService(t)
	ctx := context.Background()

	inputParams := model.TimelineParams{GroupBy: "month"}
	expectedParams := model.TimelineParams{GroupBy: "day"}

	mockRepo.On("GetIndicatorsTimeline", ctx, "camp-1", expectedParams).Return(&model.CampaignWithTimeline{}, nil)

	_, err := svc.GetIndicatorsTimeline(ctx, "camp-1", inputParams)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCampaignService_GetIndicatorsTimeline_WeekGrouping(t *testing.T) {
	svc, mockRepo, _ := setupCampaignService(t)
	ctx := context.Background()

	params := model.TimelineParams{GroupBy: "week"}

	mockRepo.On("GetIndicatorsTimeline", ctx, "camp-1", params).Return(&model.CampaignWithTimeline{
		Timeline: []model.TimelinePeriod{
			{Period: "2024-W01"},
		},
	}, nil)

	result, err := svc.GetIndicatorsTimeline(ctx, "camp-1", params)

	assert.NoError(t, err)
	assert.Len(t, result.Timeline, 1)
	mockRepo.AssertExpectations(t)
}

func TestCampaignService_GetIndicatorsTimeline_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupCampaignService(t)
	ctx := context.Background()

	params := model.TimelineParams{GroupBy: "day"}

	mockRepo.On("GetIndicatorsTimeline", ctx, "non-existent", params).Return(nil, repository.ErrNotFound)

	result, err := svc.GetIndicatorsTimeline(ctx, "non-existent", params)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestCampaignService_GetIndicatorsTimeline_UsesCache(t *testing.T) {
	svc, mockRepo, c := setupCampaignService(t)
	ctx := context.Background()

	params := model.TimelineParams{GroupBy: "day"}
	expected := &model.CampaignWithTimeline{
		Campaign: model.CampaignDetail{ID: "camp-cached"},
	}

	mockRepo.On("GetIndicatorsTimeline", ctx, "camp-cached", params).Return(expected, nil).Once()

	result1, err := svc.GetIndicatorsTimeline(ctx, "camp-cached", params)
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	result2, err := svc.GetIndicatorsTimeline(ctx, "camp-cached", params)
	require.NoError(t, err)

	assert.Equal(t, result1.Campaign.ID, result2.Campaign.ID)
	mockRepo.AssertNumberOfCalls(t, "GetIndicatorsTimeline", 1)

	c.Clear()
}

func TestCampaignService_GetIndicatorsTimeline_WithDateFilters(t *testing.T) {
	svc, mockRepo, _ := setupCampaignService(t)
	ctx := context.Background()

	params := model.TimelineParams{
		GroupBy:   "day",
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
	}

	mockRepo.On("GetIndicatorsTimeline", ctx, "camp-1", params).Return(&model.CampaignWithTimeline{
		Timeline: []model.TimelinePeriod{
			{Period: "2024-01-15"},
		},
	}, nil)

	result, err := svc.GetIndicatorsTimeline(ctx, "camp-1", params)

	assert.NoError(t, err)
	assert.Len(t, result.Timeline, 1)
	mockRepo.AssertExpectations(t)
}
