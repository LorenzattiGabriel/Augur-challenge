package handler

import (
	"database/sql"
	"net/http"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:   "healthy",
		Database: "connected",
	}

	if err := h.db.PingContext(r.Context()); err != nil {
		response.Status = "unhealthy"
		response.Database = "disconnected"
		respondJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	respondJSON(w, http.StatusOK, response)
}
