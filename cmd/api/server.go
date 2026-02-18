package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/config"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/handler"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/service"
)

type Server struct {
	cfg        *config.Config
	logger     *slog.Logger
	db         *sql.DB
	cache      *cache.Cache
	httpServer *http.Server

	indicatorHandler *handler.IndicatorHandler
	campaignHandler  *handler.CampaignHandler
	dashboardHandler *handler.DashboardHandler
	healthHandler    *handler.HealthHandler
}

func NewServer(cfg *config.Config, logger *slog.Logger, db *sql.DB, appCache *cache.Cache) *Server {
	s := &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
		cache:  appCache,
	}

	s.setupHandlers()
	s.setupHTTPServer()

	return s
}

func (s *Server) setupHandlers() {
	indicatorRepo := repository.NewIndicatorRepository(s.db)
	campaignRepo := repository.NewCampaignRepository(s.db)
	dashboardRepo := repository.NewDashboardRepository(s.db)

	indicatorService := service.NewIndicatorService(indicatorRepo, s.cache)
	campaignService := service.NewCampaignService(campaignRepo, s.cache)
	dashboardService := service.NewDashboardService(dashboardRepo, s.cache)

	s.indicatorHandler = handler.NewIndicatorHandler(indicatorService)
	s.campaignHandler = handler.NewCampaignHandler(campaignService)
	s.dashboardHandler = handler.NewDashboardHandler(dashboardService)
	s.healthHandler = handler.NewHealthHandler(s.db)
}

func (s *Server) setupHTTPServer() {
	router := s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         s.cfg.ServerHost + ":" + s.cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting server", "address", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}
