package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDashboardRouter(handler *DashboardHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/api/dashboard/summary", handler.GetSummary)
	return r
}

func TestDashboardHandler_GetSummary_Success(t *testing.T) {
	mockService := new(MockDashboardService)
	handler := NewDashboardHandler(mockService)
	r := setupDashboardRouter(handler)

	expected := &model.DashboardSummary{
		TimeRange:             "7d",
		ActiveCampaigns:       5,
		NewIndicators:         map[string]int{"ip": 100, "domain": 50},
		TopThreatActors:       []model.ThreatActorWithCount{},
		IndicatorDistribution: map[string]int{"ip": 500},
	}

	mockService.On("GetSummary", mock.Anything, "7d").Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/dashboard/summary?time_range=7d", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestDashboardHandler_GetSummary_DefaultTimeRange(t *testing.T) {
	mockService := new(MockDashboardService)
	handler := NewDashboardHandler(mockService)
	r := setupDashboardRouter(handler)

	mockService.On("GetSummary", mock.Anything, "").Return(&model.DashboardSummary{TimeRange: "7d"}, nil)

	req := httptest.NewRequest("GET", "/api/dashboard/summary", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDashboardHandler_GetSummary_InvalidTimeRange(t *testing.T) {
	mockService := new(MockDashboardService)
	handler := NewDashboardHandler(mockService)
	r := setupDashboardRouter(handler)

	req := httptest.NewRequest("GET", "/api/dashboard/summary?time_range=1y", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Equal(t, ErrCodeBadRequest, response.Error.Code)
	assert.Contains(t, response.Error.Message, "time_range")
}

func TestDashboardHandler_GetSummary_ValidTimeRanges(t *testing.T) {
	testCases := []string{"24h", "7d", "30d"}

	for _, tr := range testCases {
		t.Run("time_range_"+tr, func(t *testing.T) {
			mockService := new(MockDashboardService)
			handler := NewDashboardHandler(mockService)
			r := setupDashboardRouter(handler)

			mockService.On("GetSummary", mock.Anything, tr).Return(&model.DashboardSummary{TimeRange: tr}, nil)

			req := httptest.NewRequest("GET", "/api/dashboard/summary?time_range="+tr, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDashboardHandler_GetSummary_ServiceError(t *testing.T) {
	mockService := new(MockDashboardService)
	handler := NewDashboardHandler(mockService)
	r := setupDashboardRouter(handler)

	mockService.On("GetSummary", mock.Anything, "7d").Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/api/dashboard/summary?time_range=7d", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Equal(t, ErrCodeInternalServer, response.Error.Code)
	mockService.AssertExpectations(t)
}

func TestDashboardHandler_GetSummary_ResponseStructure(t *testing.T) {
	mockService := new(MockDashboardService)
	handler := NewDashboardHandler(mockService)
	r := setupDashboardRouter(handler)

	expected := &model.DashboardSummary{
		TimeRange:       "24h",
		ActiveCampaigns: 3,
		NewIndicators:   map[string]int{"ip": 10, "domain": 5, "url": 3, "hash": 2},
		TopThreatActors: []model.ThreatActorWithCount{
			{ThreatActor: model.ThreatActor{ID: "actor-1", Name: "APT29"}, IndicatorCount: 100},
		},
		IndicatorDistribution: map[string]int{"ip": 200, "domain": 150},
	}

	mockService.On("GetSummary", mock.Anything, "24h").Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/dashboard/summary?time_range=24h", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			TimeRange             string         `json:"time_range"`
			ActiveCampaigns       int            `json:"active_campaigns"`
			NewIndicators         map[string]int `json:"new_indicators"`
			IndicatorDistribution map[string]int `json:"indicator_distribution"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.True(t, response.Success)
	assert.Equal(t, "24h", response.Data.TimeRange)
	assert.Equal(t, 3, response.Data.ActiveCampaigns)
	assert.Equal(t, 10, response.Data.NewIndicators["ip"])
	mockService.AssertExpectations(t)
}
