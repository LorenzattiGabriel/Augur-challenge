package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/config"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/database"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/handler"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/middleware"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.NewPostgresDB(cfg.DatabaseURL())
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database migrations completed")

	appCache, err := cache.New(cache.Config{
		MaxSizeMB: cfg.CacheMaxSizeMB,
	})
	if err != nil {
		logger.Error("Failed to initialize cache", "error", err)
		os.Exit(1)
	}

	indicatorRepo := repository.NewIndicatorRepository(db)
	campaignRepo := repository.NewCampaignRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	indicatorService := service.NewIndicatorService(indicatorRepo, appCache)
	campaignService := service.NewCampaignService(campaignRepo, appCache)
	dashboardService := service.NewDashboardService(dashboardRepo, appCache)

	indicatorHandler := handler.NewIndicatorHandler(indicatorService)
	campaignHandler := handler.NewCampaignHandler(campaignService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	healthHandler := handler.NewHealthHandler(db)

	r := chi.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(cfg.RateLimitRPM))

	r.Get("/health", healthHandler.Check)

	r.Route("/api", func(r chi.Router) {
		r.Route("/indicators", func(r chi.Router) {
			r.Get("/search", indicatorHandler.Search)
			r.Get("/{id}", indicatorHandler.GetByID)
		})

		r.Route("/campaigns", func(r chi.Router) {
			r.Get("/{id}/indicators", campaignHandler.GetIndicators)
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/summary", dashboardHandler.GetSummary)
		})
	})

	server := &http.Server{
		Addr:         cfg.ServerHost + ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server stopped")
}
