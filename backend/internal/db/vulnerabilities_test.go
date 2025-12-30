package db

import (
	"context"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVulnerabilityRepository_Upsert(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	fixVersion := "1.2.3"
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-1234",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "High",
		FixVersion:      &fixVersion,
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}

	err := repo.Upsert(context.Background(), vuln)
	require.NoError(t, err)
	assert.NotZero(t, vuln.ID)
	assert.NotZero(t, vuln.CreatedAt)
}

func TestVulnerabilityRepository_GetByID(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	// Create vulnerability
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-1234",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "High",
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err := repo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(context.Background(), vuln.ID)
	require.NoError(t, err)
	assert.Equal(t, vuln.CVEID, retrieved.CVEID)
	assert.Equal(t, vuln.PackageName, retrieved.PackageName)
}

func TestVulnerabilityRepository_List(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

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
		err := repo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
	}

	// List all
	list, err := repo.List(context.Background(), 10, 0, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestVulnerabilityRepository_List_FilterBySeverity(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	// Create vulnerabilities with different severities
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
		err := repo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
	}

	// Filter by Critical
	severity := "Critical"
	list, err := repo.List(context.Background(), 10, 0, &severity, nil, nil)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Critical", list[0].Severity)
}

func TestVulnerabilityRepository_List_FilterByStatus(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	// Create vulnerabilities with different statuses
	fixDate := time.Now()
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
			Status:          "fixed",
			RemediationDate: &fixDate,
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		},
	}

	for _, vuln := range vulns {
		err := repo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
	}

	// Filter by fixed
	status := "fixed"
	list, err := repo.List(context.Background(), 10, 0, nil, &status, nil)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "fixed", list[0].Status)
}

func TestVulnerabilityRepository_Update(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	// Create vulnerability
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-1234",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "High",
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err := repo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	// Update status
	newStatus := models.StatusInProgress
	update := &models.VulnerabilityUpdateWithContext{
		Status:    &newStatus,
		UpdatedBy: "test-user",
	}

	err = repo.Update(context.Background(), vuln.ID, update)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), vuln.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusInProgress, retrieved.Status)
	assert.Equal(t, "test-user", *retrieved.UpdatedBy)
}

func TestVulnerabilityRepository_Update_CreatesHistory(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	// Create vulnerability
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-1234",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "High",
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err := repo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	// Update status
	newStatus := models.StatusFixed
	update := &models.VulnerabilityUpdateWithContext{
		Status:    &newStatus,
		UpdatedBy: "test-user",
	}

	err = repo.Update(context.Background(), vuln.ID, update)
	require.NoError(t, err)

	// Check history was created
	history, err := repo.GetHistory(context.Background(), vuln.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "status", history[0].FieldName)
	assert.Equal(t, "active", *history[0].OldValue)
	assert.Equal(t, "fixed", *history[0].NewValue)
	assert.Equal(t, "test-user", *history[0].ChangedBy)
}

func TestVulnerabilityRepository_MarkAsFixed(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

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

	var ids []int
	for _, vuln := range vulns {
		err := repo.Upsert(context.Background(), vuln)
		require.NoError(t, err)
		ids = append(ids, vuln.ID)
	}

	// Mark as fixed
	err := repo.MarkAsFixed(context.Background(), ids)
	require.NoError(t, err)

	// Verify both are marked as fixed
	for _, id := range ids {
		retrieved, err := repo.GetByID(context.Background(), id)
		require.NoError(t, err)
		assert.Equal(t, models.StatusFixed, retrieved.Status)
		assert.NotNil(t, retrieved.RemediationDate)
		assert.Equal(t, "system", *retrieved.UpdatedBy)
	}

	// Verify history was created
	history, err := repo.GetHistory(context.Background(), ids[0])
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "system", *history[0].ChangedBy)
}

func TestVulnerabilityRepository_LinkToScan(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	vulnRepo := NewVulnerabilityRepository(db)
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

	// Create vulnerability
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-1234",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "High",
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err = vulnRepo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	// Link to scan
	err = vulnRepo.LinkToScan(context.Background(), scan.ID, vuln.ID)
	require.NoError(t, err)

	// Verify link exists
	vulns, err := scanRepo.GetVulnerabilities(context.Background(), scan.ID)
	require.NoError(t, err)
	assert.Len(t, vulns, 1)
	assert.Equal(t, vuln.CVEID, vulns[0].CVEID)
}

func TestVulnerabilityRepository_GetByUniqueKey(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	// Create vulnerability
	vuln := &models.Vulnerability{
		CVEID:           "CVE-2023-1234",
		PackageName:     "openssl",
		PackageVersion:  "1.1.1",
		Severity:        "High",
		Status:          "active",
		FirstDetectedAt: time.Now(),
		LastSeenAt:      time.Now(),
	}
	err := repo.Upsert(context.Background(), vuln)
	require.NoError(t, err)

	// Get by unique key
	retrieved, err := repo.GetByUniqueKey(context.Background(), "CVE-2023-1234", "openssl", "1.1.1")
	require.NoError(t, err)
	assert.Equal(t, vuln.ID, retrieved.ID)
}

func TestVulnerabilityRepository_GetByUniqueKey_NotFound(t *testing.T) {
	db := SetupTestDatabase(t)
	defer db.Close()

	repo := NewVulnerabilityRepository(db)

	_, err := repo.GetByUniqueKey(context.Background(), "CVE-9999-9999", "notfound", "1.0.0")
	assert.Error(t, err)
}
