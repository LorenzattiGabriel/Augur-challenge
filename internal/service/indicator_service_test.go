package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIndicatorService(t *testing.T) (*IndicatorService, *MockIndicatorRepository, *cache.Cache) {
	mockRepo := new(MockIndicatorRepository)
	c, err := cache.New(cache.Config{MaxSizeMB: 10})
	require.NoError(t, err)
	svc := NewIndicatorService(mockRepo, c)
	return svc, mockRepo, c
}

func TestIndicatorService_GetByID_Success(t *testing.T) {
	svc, mockRepo, _ := setupIndicatorService(t)
	ctx := context.Background()

	expected := &model.IndicatorWithRelations{
		Indicator: model.Indicator{
			ID:         "test-uuid",
			Type:       "ip",
			Value:      "192.168.1.1",
			Confidence: 85,
		},
	}

	mockRepo.On("GetByID", ctx, "test-uuid").Return(expected, nil)

	result, err := svc.GetByID(ctx, "test-uuid")

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Value, result.Value)
	mockRepo.AssertExpectations(t)
}

func TestIndicatorService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupIndicatorService(t)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, "non-existent").Return(nil, repository.ErrNotFound)

	result, err := svc.GetByID(ctx, "non-existent")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestIndicatorService_GetByID_UsesCache(t *testing.T) {
	svc, mockRepo, c := setupIndicatorService(t)
	ctx := context.Background()

	expected := &model.IndicatorWithRelations{
		Indicator: model.Indicator{
			ID:    "cached-uuid",
			Type:  "domain",
			Value: "example.com",
		},
	}

	mockRepo.On("GetByID", ctx, "cached-uuid").Return(expected, nil).Once()

	result1, err := svc.GetByID(ctx, "cached-uuid")
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	result2, err := svc.GetByID(ctx, "cached-uuid")
	require.NoError(t, err)

	assert.Equal(t, result1.ID, result2.ID)
	mockRepo.AssertNumberOfCalls(t, "GetByID", 1)

	c.Clear()
}

func TestIndicatorService_Search_Success(t *testing.T) {
	svc, mockRepo, _ := setupIndicatorService(t)
	ctx := context.Background()

	params := model.SearchParams{
		Type:  "ip",
		Page:  1,
		Limit: 20,
	}

	expected := &model.SearchResult{
		Data:       []model.IndicatorSearchResult{{ID: "1", Type: "ip", Value: "10.0.0.1"}},
		Total:      1,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	}

	mockRepo.On("Search", ctx, params).Return(expected, nil)

	result, err := svc.Search(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Len(t, result.Data, 1)
	mockRepo.AssertExpectations(t)
}

func TestIndicatorService_Search_DefaultsPagination(t *testing.T) {
	svc, mockRepo, _ := setupIndicatorService(t)
	ctx := context.Background()

	params := model.SearchParams{
		Page:  0,
		Limit: 0,
	}

	expectedParams := model.SearchParams{
		Page:  1,
		Limit: 20,
	}

	mockRepo.On("Search", ctx, expectedParams).Return(&model.SearchResult{
		Page:  1,
		Limit: 20,
	}, nil)

	result, err := svc.Search(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.Limit)
	mockRepo.AssertExpectations(t)
}

func TestIndicatorService_Search_LimitsMaxPageSize(t *testing.T) {
	svc, mockRepo, _ := setupIndicatorService(t)
	ctx := context.Background()

	params := model.SearchParams{
		Page:  1,
		Limit: 500,
	}

	expectedParams := model.SearchParams{
		Page:  1,
		Limit: 100,
	}

	mockRepo.On("Search", ctx, expectedParams).Return(&model.SearchResult{
		Limit: 100,
	}, nil)

	result, err := svc.Search(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, 100, result.Limit)
	mockRepo.AssertExpectations(t)
}

func TestIndicatorService_Search_RepoError(t *testing.T) {
	svc, mockRepo, _ := setupIndicatorService(t)
	ctx := context.Background()

	params := model.SearchParams{Page: 1, Limit: 20}
	expectedErr := errors.New("database error")

	mockRepo.On("Search", ctx, params).Return(nil, expectedErr)

	result, err := svc.Search(ctx, params)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}
