package handler

import (
	"errors"
	"net/http"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CampaignHandler struct {
	service service.CampaignServiceInterface
}

func NewCampaignHandler(svc service.CampaignServiceInterface) *CampaignHandler {
	return &CampaignHandler{service: svc}
}

func (h *CampaignHandler) GetIndicators(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondBadRequest(w, "Campaign ID is required")
		return
	}

	if _, err := uuid.Parse(id); err != nil {
		respondBadRequest(w, "Invalid campaign ID format")
		return
	}

	params := model.TimelineParams{
		GroupBy:   r.URL.Query().Get("group_by"),
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	if params.GroupBy != "" && params.GroupBy != "day" && params.GroupBy != "week" {
		respondBadRequest(w, "Invalid group_by value. Must be 'day' or 'week'")
		return
	}

	timeline, err := h.service.GetIndicatorsTimeline(r.Context(), id, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondNotFound(w, "Campaign not found")
			return
		}
		respondInternalError(w)
		return
	}

	respondSuccess(w, timeline)
}
