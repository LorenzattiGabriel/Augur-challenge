package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IndicatorHandler struct {
	service service.IndicatorServiceInterface
}

func NewIndicatorHandler(svc service.IndicatorServiceInterface) *IndicatorHandler {
	return &IndicatorHandler{service: svc}
}

func (h *IndicatorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondBadRequest(w, "Indicator ID is required")
		return
	}

	if _, err := uuid.Parse(id); err != nil {
		respondBadRequest(w, "Invalid indicator ID format")
		return
	}

	indicator, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondNotFound(w, "Indicator not found")
			return
		}
		respondInternalError(w)
		return
	}

	respondSuccess(w, indicator)
}

func (h *IndicatorHandler) Search(w http.ResponseWriter, r *http.Request) {
	params := model.SearchParams{
		Type:           r.URL.Query().Get("type"),
		Value:          r.URL.Query().Get("value"),
		ThreatActorID:  r.URL.Query().Get("threat_actor"),
		CampaignID:     r.URL.Query().Get("campaign"),
		FirstSeenAfter: r.URL.Query().Get("first_seen_after"),
		LastSeenBefore: r.URL.Query().Get("last_seen_before"),
	}

	if params.Type != "" {
		validTypes := map[string]bool{"ip": true, "domain": true, "url": true, "hash": true}
		if !validTypes[params.Type] {
			respondBadRequest(w, "Invalid indicator type. Must be one of: ip, domain, url, hash")
			return
		}
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	params.Page = page

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	params.Limit = limit

	result, err := h.service.Search(r.Context(), params)
	if err != nil {
		respondInternalError(w)
		return
	}

	respondSuccess(w, result)
}
