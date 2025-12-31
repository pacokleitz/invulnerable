package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/invulnerable/backend/internal/analyzer"
	"github.com/invulnerable/backend/internal/api"
	"github.com/invulnerable/backend/internal/config"
	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/metrics"
	"github.com/invulnerable/backend/internal/notifier"
	"github.com/invulnerable/backend/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize database
	dbPort, err := strconv.Atoi(cfg.Database.Port)
	if err != nil {
		logger.Fatal("invalid database port", zap.String("port", cfg.Database.Port), zap.Error(err))
	}

	database, err := db.New(db.Config{
		Host:     cfg.Database.Host,
		Port:     dbPort,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("connected to database successfully")

	// Initialize S3 client for SBOM storage
	s3Client, err := createS3Client(cfg.S3)
	if err != nil {
		logger.Fatal("failed to create S3 client", zap.Error(err))
	}
	s3Storage := storage.NewS3Storage(s3Client, cfg.S3.Bucket)
	logger.Info("initialized S3 storage",
		zap.String("endpoint", cfg.S3.Endpoint),
		zap.String("bucket", cfg.S3.Bucket))

	// Initialize repositories
	imageRepo := db.NewImageRepository(database)
	scanRepo := db.NewScanRepository(database)
	vulnRepo := db.NewVulnerabilityRepository(database)
	sbomRepo := db.NewSBOMRepository(database, s3Storage)

	// Initialize services
	analyzerSvc := analyzer.New(scanRepo, vulnRepo)
	metricsSvc := metrics.New(database)
	frontendURL := getEnv("FRONTEND_URL", "")
	notifierSvc := notifier.New(logger, frontendURL)

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
	port := cfg.Server.Port
	go func() {
		logger.Info("starting server", zap.String("port", port))
		if err := e.Start(":" + port); err != nil {
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

// createS3Client creates an AWS S3 client with custom endpoint support
func createS3Client(s3Config config.S3Config) (*s3.Client, error) {
	// Create custom resolver for endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               s3Config.Endpoint,
					HostnameImmutable: true,
					SigningRegion:     s3Config.Region,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)

	// Load AWS config with custom credentials
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(s3Config.Region),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				s3Config.AccessKey,
				s3Config.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with path-style addressing (required for MinIO)
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
