package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDashboardService(t *testing.T) (*DashboardService, *MockDashboardRepository, *cache.Cache) {
	mockRepo := new(MockDashboardRepository)
	c, err := cache.New(cache.Config{MaxSizeMB: 10})
	require.NoError(t, err)
	svc := NewDashboardService(mockRepo, c)
	return svc, mockRepo, c
}

func TestDashboardService_GetSummary_Success(t *testing.T) {
	svc, mockRepo, _ := setupDashboardService(t)
	ctx := context.Background()

	expected := &model.DashboardSummary{
		TimeRange:       "7d",
		ActiveCampaigns: 5,
		NewIndicators:   map[string]int{"ip": 100, "domain": 50},
		TopThreatActors: []model.ThreatActorWithCount{
			{ThreatActor: model.ThreatActor{ID: "actor-1", Name: "APT29"}, IndicatorCount: 200},
		},
		IndicatorDistribution: map[string]int{"ip": 500, "domain": 300},
	}

	mockRepo.On("GetSummary", ctx, "7d").Return(expected, nil)

	result, err := svc.GetSummary(ctx, "7d")

	assert.NoError(t, err)
	assert.Equal(t, "7d", result.TimeRange)
	assert.Equal(t, 5, result.ActiveCampaigns)
	assert.Equal(t, 100, result.NewIndicators["ip"])
	assert.Len(t, result.TopThreatActors, 1)
	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_DefaultsTimeRange(t *testing.T) {
	svc, mockRepo, _ := setupDashboardService(t)
	ctx := context.Background()

	mockRepo.On("GetSummary", ctx, "7d").Return(&model.DashboardSummary{TimeRange: "7d"}, nil)

	result, err := svc.GetSummary(ctx, "")

	assert.NoError(t, err)
	assert.Equal(t, "7d", result.TimeRange)
	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_InvalidTimeRangeDefaultsTo7d(t *testing.T) {
	svc, mockRepo, _ := setupDashboardService(t)
	ctx := context.Background()

	mockRepo.On("GetSummary", ctx, "7d").Return(&model.DashboardSummary{TimeRange: "7d"}, nil)

	result, err := svc.GetSummary(ctx, "invalid")

	assert.NoError(t, err)
	assert.Equal(t, "7d", result.TimeRange)
	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_ValidTimeRanges(t *testing.T) {
	testCases := []string{"24h", "7d", "30d"}

	for _, tr := range testCases {
		t.Run("time_range_"+tr, func(t *testing.T) {
			svc, mockRepo, _ := setupDashboardService(t)
			ctx := context.Background()

			mockRepo.On("GetSummary", ctx, tr).Return(&model.DashboardSummary{TimeRange: tr}, nil)

			result, err := svc.GetSummary(ctx, tr)

			assert.NoError(t, err)
			assert.Equal(t, tr, result.TimeRange)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDashboardService_GetSummary_UsesCache(t *testing.T) {
	svc, mockRepo, c := setupDashboardService(t)
	ctx := context.Background()

	expected := &model.DashboardSummary{
		TimeRange:       "24h",
		ActiveCampaigns: 3,
	}

	mockRepo.On("GetSummary", ctx, "24h").Return(expected, nil).Once()

	result1, err := svc.GetSummary(ctx, "24h")
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	result2, err := svc.GetSummary(ctx, "24h")
	require.NoError(t, err)

	assert.Equal(t, result1.ActiveCampaigns, result2.ActiveCampaigns)
	mockRepo.AssertNumberOfCalls(t, "GetSummary", 1)

	c.Clear()
}

func TestDashboardService_GetSummary_RepoError(t *testing.T) {
	svc, mockRepo, _ := setupDashboardService(t)
	ctx := context.Background()

	expectedErr := errors.New("database connection failed")
	mockRepo.On("GetSummary", ctx, "7d").Return(nil, expectedErr)

	result, err := svc.GetSummary(ctx, "7d")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, "database connection failed", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_DifferentCacheKeys(t *testing.T) {
	svc, mockRepo, c := setupDashboardService(t)
	ctx := context.Background()

	mockRepo.On("GetSummary", ctx, "24h").Return(&model.DashboardSummary{TimeRange: "24h"}, nil).Once()
	mockRepo.On("GetSummary", ctx, "7d").Return(&model.DashboardSummary{TimeRange: "7d"}, nil).Once()

	_, err := svc.GetSummary(ctx, "24h")
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	_, err = svc.GetSummary(ctx, "7d")
	require.NoError(t, err)

	mockRepo.AssertNumberOfCalls(t, "GetSummary", 2)

	c.Clear()
}
