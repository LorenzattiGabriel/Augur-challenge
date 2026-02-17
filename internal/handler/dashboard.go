package handler

import (
	"net/http"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: svc}
}

func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("time_range")

	if timeRange != "" {
		validRanges := map[string]bool{"24h": true, "7d": true, "30d": true}
		if !validRanges[timeRange] {
			respondBadRequest(w, "Invalid time_range. Must be one of: 24h, 7d, 30d")
			return
		}
	}

	summary, err := h.service.GetSummary(r.Context(), timeRange)
	if err != nil {
		respondInternalError(w)
		return
	}

	respondSuccess(w, summary)
}
