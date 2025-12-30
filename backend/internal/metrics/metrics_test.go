package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDashboardMetrics(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	service := New(database)
	imageRepo := db.NewImageRepository(database)
	scanRepo := db.NewScanRepository(database)
	vulnRepo := db.NewVulnerabilityRepository(database)

	// Create test data
	// Image 1
	image1 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image1)
	require.NoError(t, err)

	// Image 2
	image2 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/postgres",
		Tag:        "15",
	}
	err = imageRepo.Create(context.Background(), image2)
	require.NoError(t, err)

	// Recent scan for image 1
	scan1 := &models.Scan{
		ImageID:     image1.ID,
		ScanDate:    time.Now(),
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = scanRepo.Create(context.Background(), scan1)
	require.NoError(t, err)

	// Old scan for image 2 (more than 24 hours ago)
	scan2 := &models.Scan{
		ImageID:     image2.ID,
		ScanDate:    time.Now().Add(-48 * time.Hour),
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = scanRepo.Create(context.Background(), scan2)
	require.NoError(t, err)

	// Create vulnerabilities
	vulns := []*models.Vulnerability{
		{
			CVEID:           "CVE-2023-0001",
			PackageName:     "openssl",
			PackageVersion:  "1.1.1",
			Severity:        "Critical",
			Status:          "active",
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
		{
			CVEID:           "CVE-2023-0002",
			PackageName:     "curl",
			PackageVersion:  "7.74.0",
			Severity:        "High",
			Status:          "active",
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
		{
			CVEID:           "CVE-2023-0003",
			PackageName:     "libxml2",
			PackageVersion:  "2.9.0",
			Severity:        "Medium",
			Status:          "active",
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
		{
			CVEID:           "CVE-2023-0004",
			PackageName:     "zlib",
			PackageVersion:  "1.2.8",
			Severity:        "Low",
			Status:          "active",
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
	}

	for _, vuln := range vulns {
		err = vulnRepo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
	}

	// Get metrics
	metrics, err := service.GetDashboardMetrics(context.Background(), nil)
	require.NoError(t, err)

	// Assertions
	assert.Equal(t, 2, metrics.TotalImages)
	assert.Equal(t, 2, metrics.TotalScans)
	assert.Equal(t, 4, metrics.TotalVulnerabilities)
	assert.Equal(t, 4, metrics.ActiveVulnerabilities)
	assert.Equal(t, 1, metrics.SeverityCounts.Critical)
	assert.Equal(t, 1, metrics.SeverityCounts.High)
	assert.Equal(t, 1, metrics.SeverityCounts.Medium)
	assert.Equal(t, 1, metrics.SeverityCounts.Low)
	assert.Equal(t, 1, metrics.RecentScans) // Only scan1 is within 24 hours
}

func TestGetDashboardMetrics_WithFixFilter(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	service := New(database)
	vulnRepo := db.NewVulnerabilityRepository(database)

	// Create vulnerabilities with and without fixes
	fixVersion := "1.2.0"
	vulns := []*models.Vulnerability{
		{
			CVEID:           "CVE-2023-0001",
			PackageName:     "openssl",
			PackageVersion:  "1.1.1",
			Severity:        "Critical",
			Status:          "active",
			FixVersion:      &fixVersion,
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
		{
			CVEID:           "CVE-2023-0002",
			PackageName:     "curl",
			PackageVersion:  "7.74.0",
			Severity:        "High",
			Status:          "active",
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
	}

	for _, vuln := range vulns {
		err := vulnRepo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
	}

	// Get metrics with hasFix=true
	hasFix := true
	metrics, err := service.GetDashboardMetrics(context.Background(), &hasFix)
	require.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalVulnerabilities)
	assert.Equal(t, 1, metrics.ActiveVulnerabilities)
	assert.Equal(t, 1, metrics.SeverityCounts.Critical)
	assert.Equal(t, 0, metrics.SeverityCounts.High)

	// Get metrics with hasFix=false
	hasFix = false
	metrics, err = service.GetDashboardMetrics(context.Background(), &hasFix)
	require.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalVulnerabilities)
	assert.Equal(t, 1, metrics.ActiveVulnerabilities)
	assert.Equal(t, 0, metrics.SeverityCounts.Critical)
	assert.Equal(t, 1, metrics.SeverityCounts.High)
}

func TestGetDashboardMetrics_EmptyDatabase(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	service := New(database)

	metrics, err := service.GetDashboardMetrics(context.Background(), nil)
	require.NoError(t, err)

	// All metrics should be zero
	assert.Equal(t, 0, metrics.TotalImages)
	assert.Equal(t, 0, metrics.TotalScans)
	assert.Equal(t, 0, metrics.TotalVulnerabilities)
	assert.Equal(t, 0, metrics.ActiveVulnerabilities)
	assert.Equal(t, 0, metrics.SeverityCounts.Critical)
	assert.Equal(t, 0, metrics.SeverityCounts.High)
	assert.Equal(t, 0, metrics.SeverityCounts.Medium)
	assert.Equal(t, 0, metrics.SeverityCounts.Low)
	assert.Equal(t, 0, metrics.RecentScans)
}
