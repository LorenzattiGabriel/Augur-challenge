package main

import (
	"github.com/LorenzattiGabriel/threat-intel-api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *Server) setupRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recovery(s.logger))
	r.Use(middleware.Logger(s.logger))
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(s.cfg.RateLimitRPM))

	r.Get("/health", s.healthHandler.Check)

	r.Route("/api", func(r chi.Router) {
		r.Route("/indicators", func(r chi.Router) {
			r.Get("/search", s.indicatorHandler.Search)
			r.Get("/{id}", s.indicatorHandler.GetByID)
		})

		r.Route("/campaigns", func(r chi.Router) {
			r.Get("/{id}/indicators", s.campaignHandler.GetIndicators)
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/summary", s.dashboardHandler.GetSummary)
		})
	})

	return r
}
