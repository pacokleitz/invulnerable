package db

import (
	"context"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanRepository_Create(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

	// Create image first
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image)
	require.NoError(t, err)

	// Create scan
	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"
	scan := &models.Scan{
		ImageID:      image.ID,
		ScanDate:     time.Now(),
		SyftVersion:  &syftVersion,
		GrypeVersion: &grypeVersion,
		Status:       "completed",
		SLACritical:  7,
		SLAHigh:      30,
		SLAMedium:    90,
		SLALow:       180,
	}

	err = repo.Create(context.Background(), scan)
	require.NoError(t, err)
	assert.NotZero(t, scan.ID)
	assert.NotZero(t, scan.CreatedAt)
	assert.NotZero(t, scan.UpdatedAt)
}

func TestScanRepository_GetWithDetails(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)
	vulnRepo := NewVulnerabilityRepository(db)

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
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = repo.Create(context.Background(), scan)
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
	}

	for _, vuln := range vulns {
		err = vulnRepo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
		err = vulnRepo.LinkToScan(context.Background(), scan.ID, vuln.ID)
		require.NoError(t, err)
	}

	// Get scan with details
	retrieved, err := repo.GetWithDetails(context.Background(), scan.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, scan.ID, retrieved.ID)
	assert.Equal(t, "docker.io/library/nginx:latest", retrieved.ImageName)
	assert.Equal(t, 2, retrieved.VulnerabilityCount)
	assert.Equal(t, 1, retrieved.CriticalCount)
	assert.Equal(t, 1, retrieved.HighCount)
}

func TestScanRepository_List(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

	// Create images
	image1 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image1)
	require.NoError(t, err)

	image2 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/postgres",
		Tag:        "15",
	}
	err = imageRepo.Create(context.Background(), image2)
	require.NoError(t, err)

	// Create scans
	for _, imgID := range []int{image1.ID, image2.ID} {
		scan := &models.Scan{
			ImageID:     imgID,
			Status:      "completed",
			SLACritical: 7,
			SLAHigh:     30,
			SLAMedium:   90,
			SLALow:      180,
		}
		err = repo.Create(context.Background(), scan)
		require.NoError(t, err)
	}

	// List scans
	scans, err := repo.List(context.Background(), 10, 0, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, scans, 2)
}

func TestScanRepository_List_FilterByImage(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

	// Create two images
	image1 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image1)
	require.NoError(t, err)

	image2 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/alpine",
		Tag:        "3.18",
	}
	err = imageRepo.Create(context.Background(), image2)
	require.NoError(t, err)

	// Create scans for both images
	for _, imgID := range []int{image1.ID, image2.ID} {
		scan := &models.Scan{
			ImageID:     imgID,
			Status:      "completed",
			SLACritical: 7,
			SLAHigh:     30,
			SLAMedium:   90,
			SLALow:      180,
		}
		err = repo.Create(context.Background(), scan)
		require.NoError(t, err)
	}

	// Filter by image1
	scans, err := repo.List(context.Background(), 10, 0, &image1.ID, nil, nil)
	require.NoError(t, err)
	assert.Len(t, scans, 1)
	assert.Equal(t, image1.ID, scans[0].ImageID)
}

func TestScanRepository_GetVulnerabilities(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)
	vulnRepo := NewVulnerabilityRepository(db)

	// Create image and scan
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image)
	require.NoError(t, err)

	scan := &models.Scan{
		ImageID:     image.ID,
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = repo.Create(context.Background(), scan)
	require.NoError(t, err)

	// Create vulnerabilities
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
	err = vulnRepo.LinkToScan(context.Background(), scan.ID, vuln.ID)
	require.NoError(t, err)

	// Get vulnerabilities for scan
	vulns, err := repo.GetVulnerabilities(context.Background(), scan.ID)
	require.NoError(t, err)
	assert.Len(t, vulns, 1)
	assert.Equal(t, "CVE-2023-0001", vulns[0].CVEID)
}

func TestScanRepository_GetPreviousScan(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

	// Create image
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image)
	require.NoError(t, err)

	// Create two scans with explicit scan dates
	now := time.Now()
	scan1 := &models.Scan{
		ImageID:     image.ID,
		ScanDate:    now.Add(-2 * time.Hour), // 2 hours ago
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = repo.Create(context.Background(), scan1)
	require.NoError(t, err)

	scan2 := &models.Scan{
		ImageID:     image.ID,
		ScanDate:    now.Add(-1 * time.Hour), // 1 hour ago
		Status:      "completed",
		SLACritical: 7,
		SLAHigh:     30,
		SLAMedium:   90,
		SLALow:      180,
	}
	err = repo.Create(context.Background(), scan2)
	require.NoError(t, err)

	// Get previous scan (using RFC3339 format with timezone)
	previous, err := repo.GetPreviousScan(context.Background(), image.ID, scan2.ScanDate.Format(time.RFC3339))
	require.NoError(t, err)
	assert.Equal(t, scan1.ID, previous.ID)
}
