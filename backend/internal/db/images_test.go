package db

import (
	"context"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageRepository_Create(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)

	digest := "sha256:abc123"
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
		Digest:     &digest,
	}

	err := repo.Create(context.Background(), image)
	require.NoError(t, err)
	assert.NotZero(t, image.ID)
	assert.NotZero(t, image.CreatedAt)
	assert.NotZero(t, image.UpdatedAt)
}

func TestImageRepository_Create_UpdateDigest(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)

	// Create image with initial digest
	digest1 := "sha256:abc123"
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
		Digest:     &digest1,
	}

	err := repo.Create(context.Background(), image)
	require.NoError(t, err)
	firstID := image.ID

	// Create same image with different digest (should update)
	digest2 := "sha256:def456"
	image2 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
		Digest:     &digest2,
	}

	err = repo.Create(context.Background(), image2)
	require.NoError(t, err)

	// Should have same ID (updated, not inserted)
	assert.Equal(t, firstID, image2.ID)
	assert.Equal(t, &digest2, image2.Digest)
}

func TestImageRepository_GetByID(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)

	// Create image first
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/alpine",
		Tag:        "3.18",
	}
	err := repo.Create(context.Background(), image)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(context.Background(), image.ID)
	require.NoError(t, err)
	assert.Equal(t, image.Registry, retrieved.Registry)
	assert.Equal(t, image.Repository, retrieved.Repository)
	assert.Equal(t, image.Tag, retrieved.Tag)
}

func TestImageRepository_List(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)
	scanRepo := NewScanRepository(db)
	vulnRepo := NewVulnerabilityRepository(db)

	// Create images
	image1 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := repo.Create(context.Background(), image1)
	require.NoError(t, err)

	image2 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/alpine",
		Tag:        "3.18",
	}
	err = repo.Create(context.Background(), image2)
	require.NoError(t, err)

	// Create scan for image1
	syft := "0.100.0"
	grype := "0.74.0"
	scan := &models.Scan{
		ImageID:      image1.ID,
		ScanDate:     time.Now(),
		SyftVersion:  &syft,
		GrypeVersion: &grype,
		Status:       "completed",
		SLACritical:  7,
		SLAHigh:      30,
		SLAMedium:    90,
		SLALow:       180,
	}
	err = scanRepo.Create(context.Background(), scan)
	require.NoError(t, err)

	// Create vulnerabilities for image1
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

	// List images
	images, err := repo.List(context.Background(), 10, 0, nil)
	require.NoError(t, err)
	assert.Len(t, images, 2)

	// Find image1 and check stats
	var img1Stats *models.ImageWithStats
	for i := range images {
		if images[i].ID == image1.ID {
			img1Stats = &images[i]
			break
		}
	}
	require.NotNil(t, img1Stats)
	assert.Equal(t, 1, img1Stats.ScanCount)
	assert.Equal(t, 1, img1Stats.CriticalCount)
}

func TestImageRepository_GetByName(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)

	// Create image
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := repo.Create(context.Background(), image)
	require.NoError(t, err)

	// Get by name
	retrieved, err := repo.GetByName(context.Background(), "docker.io", "library/nginx", "latest")
	require.NoError(t, err)
	assert.Equal(t, image.ID, retrieved.ID)
	assert.Equal(t, image.Registry, retrieved.Registry)
}

func TestImageRepository_GetByName_NotFound(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)

	_, err := repo.GetByName(context.Background(), "docker.io", "library/notfound", "latest")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestImageRepository_GetScanHistory(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)
	scanRepo := NewScanRepository(db)

	// Create image
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := repo.Create(context.Background(), image)
	require.NoError(t, err)

	// Create scans
	syft := "0.100.0"
	grype := "0.74.0"
	for i := 0; i < 3; i++ {
		scan := &models.Scan{
			ImageID:      image.ID,
			ScanDate:     time.Now().Add(time.Duration(-i) * time.Hour),
			SyftVersion:  &syft,
			GrypeVersion: &grype,
			Status:       "completed",
			SLACritical:  7,
			SLAHigh:      30,
			SLAMedium:    90,
			SLALow:       180,
		}
		err = scanRepo.Create(context.Background(), scan)
		require.NoError(t, err)
	}

	// Get scan history
	scans, err := repo.GetScanHistory(context.Background(), image.ID, 10, nil)
	require.NoError(t, err)
	assert.Len(t, scans, 3)

	// Should be ordered by scan_date DESC
	assert.True(t, scans[0].ScanDate.After(scans[1].ScanDate))
	assert.True(t, scans[1].ScanDate.After(scans[2].ScanDate))
}

func TestImageRepository_GetScanHistory_EmptyResult(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewImageRepository(db)

	// Image doesn't exist
	scans, err := repo.GetScanHistory(context.Background(), 999, 10, nil)
	require.NoError(t, err)
	assert.Len(t, scans, 0)
}
