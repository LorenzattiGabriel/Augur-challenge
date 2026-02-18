package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCampaignRouter(handler *CampaignHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/api/campaigns/{id}/indicators", handler.GetIndicators)
	return r
}

func TestCampaignHandler_GetIndicators_Success(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	expected := &model.CampaignWithTimeline{
		Campaign: model.CampaignDetail{
			ID:     "550e8400-e29b-41d4-a716-446655440000",
			Name:   "Operation Test",
			Status: "active",
		},
		Timeline: []model.TimelinePeriod{
			{Period: "2024-01-01", Counts: map[string]int{"ip": 5}},
		},
		Summary: model.TimelineSummary{TotalIndicators: 5},
	}

	mockService.On("GetIndicatorsTimeline", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", model.TimelineParams{}).Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/campaigns/550e8400-e29b-41d4-a716-446655440000/indicators", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestCampaignHandler_GetIndicators_InvalidUUID(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	req := httptest.NewRequest("GET", "/api/campaigns/invalid-uuid/indicators", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Equal(t, ErrCodeBadRequest, response.Error.Code)
	assert.Contains(t, response.Error.Message, "Invalid campaign ID format")
}

func TestCampaignHandler_GetIndicators_NotFound(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	mockService.On("GetIndicatorsTimeline", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", model.TimelineParams{}).Return(nil, repository.ErrNotFound)

	req := httptest.NewRequest("GET", "/api/campaigns/550e8400-e29b-41d4-a716-446655440000/indicators", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Equal(t, ErrCodeNotFound, response.Error.Code)
	mockService.AssertExpectations(t)
}

func TestCampaignHandler_GetIndicators_InvalidGroupBy(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	req := httptest.NewRequest("GET", "/api/campaigns/550e8400-e29b-41d4-a716-446655440000/indicators?group_by=month", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error.Message, "group_by")
}

func TestCampaignHandler_GetIndicators_ValidGroupByDay(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	expectedParams := model.TimelineParams{GroupBy: "day"}
	mockService.On("GetIndicatorsTimeline", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", expectedParams).Return(&model.CampaignWithTimeline{}, nil)

	req := httptest.NewRequest("GET", "/api/campaigns/550e8400-e29b-41d4-a716-446655440000/indicators?group_by=day", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCampaignHandler_GetIndicators_ValidGroupByWeek(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	expectedParams := model.TimelineParams{GroupBy: "week"}
	mockService.On("GetIndicatorsTimeline", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", expectedParams).Return(&model.CampaignWithTimeline{}, nil)

	req := httptest.NewRequest("GET", "/api/campaigns/550e8400-e29b-41d4-a716-446655440000/indicators?group_by=week", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCampaignHandler_GetIndicators_WithDateFilters(t *testing.T) {
	mockService := new(MockCampaignService)
	handler := NewCampaignHandler(mockService)
	r := setupCampaignRouter(handler)

	expectedParams := model.TimelineParams{
		GroupBy:   "day",
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
	}
	mockService.On("GetIndicatorsTimeline", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", expectedParams).Return(&model.CampaignWithTimeline{}, nil)

	req := httptest.NewRequest("GET", "/api/campaigns/550e8400-e29b-41d4-a716-446655440000/indicators?group_by=day&start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
