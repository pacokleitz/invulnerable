package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/metrics"
	"github.com/invulnerable/backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetricsHandler_GetMetrics(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	metricsService := metrics.New(database)
	handler := NewMetricsHandler(logger, metricsService)

	// Create test data
	imageRepo := db.NewImageRepository(database)
	scanRepo := db.NewScanRepository(database)
	vulnRepo := db.NewVulnerabilityRepository(database)

	// Create image
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image)
	require.NoError(t, err)

	// Create scan
	scan := &models.Scan{
		ImageID:     image.ID,
		ScanDate:    time.Now(),
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = scanRepo.Create(context.Background(), scan)
	require.NoError(t, err)

	// Create vulnerability
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-0001",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "Critical",
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err = vulnRepo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.GetMetrics(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total_images":1`)
	assert.Contains(t, rec.Body.String(), `"total_scans":1`)
	assert.Contains(t, rec.Body.String(), `"total_vulnerabilities":1`)
}

func TestMetricsHandler_GetMetrics_WithHasFixFilter(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	metricsService := metrics.New(database)
	handler := NewMetricsHandler(logger, metricsService)

	// Create vulnerability with fix
	vulnRepo := db.NewVulnerabilityRepository(database)
	fixVersion := "1.2.0"
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-0001",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "Critical",
		Status:          "active",
		FixVersion:      &fixVersion,
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err := vulnRepo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics?has_fix=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.GetMetrics(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total_vulnerabilities":1`)
}

func TestMetricsHandler_GetMetrics_InvalidHasFix(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	metricsService := metrics.New(database)
	handler := NewMetricsHandler(logger, metricsService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics?has_fix=notabool", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetMetrics(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "invalid has_fix parameter")
}

func TestMetricsHandler_GetMetrics_EmptyDatabase(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	metricsService := metrics.New(database)
	handler := NewMetricsHandler(logger, metricsService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetMetrics(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total_images":0`)
	assert.Contains(t, rec.Body.String(), `"total_scans":0`)
	assert.Contains(t, rec.Body.String(), `"total_vulnerabilities":0`)
}
