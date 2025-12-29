package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/invulnerable/backend/internal/analyzer"
	"github.com/invulnerable/backend/internal/api"
	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/metrics"
	"github.com/invulnerable/backend/internal/notifier"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Load configuration from environment
	cfg := loadConfig()

	// Initialize database
	database, err := db.New(cfg.DB)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("connected to database successfully")

	// Initialize repositories
	imageRepo := db.NewImageRepository(database)
	scanRepo := db.NewScanRepository(database)
	vulnRepo := db.NewVulnerabilityRepository(database)
	sbomRepo := db.NewSBOMRepository(database)

	// Initialize services
	analyzerSvc := analyzer.New(scanRepo, vulnRepo)
	metricsSvc := metrics.New(database)
	notifierSvc := notifier.New(logger, cfg.FrontendURL)

	// Initialize handlers
	healthHandler := api.NewHealthHandler(database)
	scanHandler := api.NewScanHandler(logger, imageRepo, scanRepo, vulnRepo, sbomRepo, analyzerSvc, notifierSvc)
	vulnHandler := api.NewVulnerabilityHandler(logger, vulnRepo)
	imageHandler := api.NewImageHandler(logger, imageRepo)
	metricsHandler := api.NewMetricsHandler(logger, metricsSvc)
	userHandler := api.NewUserHandler(logger)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("request",
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.Duration("latency", v.Latency),
				)
			} else {
				logger.Error("request error",
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.Error(v.Error),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health endpoints
	e.GET("/health", healthHandler.Health)
	e.GET("/ready", healthHandler.Ready)

	// API routes
	api := e.Group("/api/v1")

	// Scans
	api.POST("/scans", scanHandler.CreateScan)
	api.GET("/scans", scanHandler.ListScans)
	api.GET("/scans/:id", scanHandler.GetScan)
	api.GET("/scans/:id/sbom", scanHandler.GetSBOM)
	api.GET("/scans/:id/diff", scanHandler.GetScanDiff)

	// Vulnerabilities
	api.GET("/vulnerabilities", vulnHandler.ListVulnerabilities)
	api.GET("/vulnerabilities/:cve", vulnHandler.GetVulnerabilityByCVE)
	api.PATCH("/vulnerabilities/:id", vulnHandler.UpdateVulnerability)
	api.PATCH("/vulnerabilities/bulk", vulnHandler.BulkUpdateVulnerabilities)
	api.GET("/vulnerabilities/:id/history", vulnHandler.GetVulnerabilityHistory)

	// Images
	api.GET("/images", imageHandler.ListImages)
	api.GET("/images/:id/history", imageHandler.GetImageHistory)

	// Metrics
	api.GET("/metrics", metricsHandler.GetMetrics)

	// User
	api.GET("/user/me", userHandler.GetCurrentUser)

	// Start server
	go func() {
		logger.Info("starting server", zap.String("port", cfg.Port))
		if err := e.Start(":" + cfg.Port); err != nil {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed", zap.Error(err))
	}

	logger.Info("server stopped gracefully")
}

type Config struct {
	Port        string
	FrontendURL string
	DB          db.Config
}

func loadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", ""),
		DB: db.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "invulnerable"),
			Password: getEnv("DB_PASSWORD", "invulnerable"),
			DBName:   getEnv("DB_NAME", "invulnerable"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
