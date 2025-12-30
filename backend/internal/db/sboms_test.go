package db

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSBOMRepository_Create(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewSBOMRepository(db)
	scanRepo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

	// Create image and scan first
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
	err = scanRepo.Create(context.Background(), scan)
	require.NoError(t, err)

	// Create SBOM
	version := "1.5"
	document := json.RawMessage(`{"bomFormat":"CycloneDX","specVersion":"1.5"}`)
	sbom := &models.SBOM{
		ScanID:   scan.ID,
		Format:   "cyclonedx",
		Version:  &version,
		Document: document,
	}

	err = repo.Create(context.Background(), sbom)
	require.NoError(t, err)
	assert.NotZero(t, sbom.ID)
	assert.NotZero(t, sbom.CreatedAt)
}

func TestSBOMRepository_GetByScanID(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewSBOMRepository(db)
	scanRepo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

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
	err = scanRepo.Create(context.Background(), scan)
	require.NoError(t, err)

	// Create SBOM
	version := "1.5"
	document := json.RawMessage(`{"bomFormat":"CycloneDX"}`)
	sbom := &models.SBOM{
		ScanID:   scan.ID,
		Format:   "cyclonedx",
		Version:  &version,
		Document: document,
	}
	err = repo.Create(context.Background(), sbom)
	require.NoError(t, err)

	// Get by scan ID
	retrieved, err := repo.GetByScanID(context.Background(), scan.ID)
	require.NoError(t, err)
	assert.Equal(t, sbom.ID, retrieved.ID)
	assert.Equal(t, sbom.Format, retrieved.Format)
}

func TestSBOMRepository_GetByScanID_NotFound(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewSBOMRepository(db)

	_, err := repo.GetByScanID(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSBOMRepository_GetDocumentByScanID(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewSBOMRepository(db)
	scanRepo := NewScanRepository(db)
	imageRepo := NewImageRepository(db)

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
	err = scanRepo.Create(context.Background(), scan)
	require.NoError(t, err)

	// Create SBOM
	document := json.RawMessage(`{"bomFormat":"CycloneDX","specVersion":"1.5"}`)
	sbom := &models.SBOM{
		ScanID:   scan.ID,
		Format:   "cyclonedx",
		Document: document,
	}
	err = repo.Create(context.Background(), sbom)
	require.NoError(t, err)

	// Get document
	doc, err := repo.GetDocumentByScanID(context.Background(), scan.ID)
	require.NoError(t, err)
	assert.JSONEq(t, string(document), string(doc))
}
