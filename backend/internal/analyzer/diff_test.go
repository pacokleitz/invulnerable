package analyzer

import (
	"context"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repositories
type MockScanRepo struct {
	mock.Mock
}

func (m *MockScanRepo) GetByID(ctx context.Context, id int) (*models.Scan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Scan), args.Error(1)
}

func (m *MockScanRepo) GetPreviousScan(ctx context.Context, imageID int, currentScanDate string) (*models.Scan, error) {
	args := m.Called(ctx, imageID, currentScanDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Scan), args.Error(1)
}

func (m *MockScanRepo) GetVulnerabilities(ctx context.Context, scanID int) ([]models.Vulnerability, error) {
	args := m.Called(ctx, scanID)
	return args.Get(0).([]models.Vulnerability), args.Error(1)
}

type MockVulnRepo struct {
	mock.Mock
}

func (m *MockVulnRepo) MarkAsFixed(ctx context.Context, vulnerabilityIDs []int) error {
	args := m.Called(ctx, vulnerabilityIDs)
	return args.Error(0)
}

func TestMakeVulnKey(t *testing.T) {
	vuln := models.Vulnerability{
		CVEID:          "CVE-2023-1234",
		PackageName:    "openssl",
		PackageVersion: "1.1.1",
	}

	key := makeVulnKey(vuln)
	expected := "CVE-2023-1234:openssl:1.1.1"
	assert.Equal(t, expected, key)
}

func TestAnalyzer_CompareScan_NoPreviousScan(t *testing.T) {
	mockScanRepo := new(MockScanRepo)
	mockVulnRepo := new(MockVulnRepo)
	analyzer := New(mockScanRepo, mockVulnRepo)

	ctx := context.Background()
	scanID := 1
	imageID := 100
	now := time.Now()

	currentScan := &models.Scan{
		ID:       scanID,
		ImageID:  imageID,
		ScanDate: now,
	}

	currentVulns := []models.Vulnerability{
		{
			ID:             1,
			CVEID:          "CVE-2023-1",
			PackageName:    "pkg1",
			PackageVersion: "1.0",
			Severity:       "High",
		},
		{
			ID:             2,
			CVEID:          "CVE-2023-2",
			PackageName:    "pkg2",
			PackageVersion: "2.0",
			Severity:       "Medium",
		},
	}

	mockScanRepo.On("GetByID", ctx, scanID).Return(currentScan, nil)
	mockScanRepo.On("GetPreviousScan", ctx, imageID, mock.Anything).Return(nil, nil)
	mockScanRepo.On("GetVulnerabilities", ctx, scanID).Return(currentVulns, nil)

	diff, err := analyzer.CompareScan(ctx, scanID)

	require.NoError(t, err)
	assert.Equal(t, scanID, diff.ScanID)
	assert.Equal(t, 0, diff.PreviousScanID)
	assert.Len(t, diff.NewVulns, 2)
	assert.Len(t, diff.FixedVulns, 0)
	assert.Len(t, diff.PersistentVulns, 0)
	assert.Equal(t, 2, diff.Summary.NewCount)
	assert.Equal(t, 0, diff.Summary.FixedCount)
	assert.Equal(t, 0, diff.Summary.PersistentCount)

	mockScanRepo.AssertExpectations(t)
}

func TestAnalyzer_CompareScan_WithPreviousScan(t *testing.T) {
	mockScanRepo := new(MockScanRepo)
	mockVulnRepo := new(MockVulnRepo)
	analyzer := New(mockScanRepo, mockVulnRepo)

	ctx := context.Background()
	scanID := 2
	previousScanID := 1
	imageID := 100
	now := time.Now()

	currentScan := &models.Scan{
		ID:       scanID,
		ImageID:  imageID,
		ScanDate: now,
	}

	previousScan := &models.Scan{
		ID:       previousScanID,
		ImageID:  imageID,
		ScanDate: now.Add(-24 * time.Hour),
	}

	// Current scan has: CVE-1 (new), CVE-2 (persistent), CVE-3 (new)
	currentVulns := []models.Vulnerability{
		{ID: 1, CVEID: "CVE-2023-1", PackageName: "pkg1", PackageVersion: "1.0"},
		{ID: 2, CVEID: "CVE-2023-2", PackageName: "pkg2", PackageVersion: "2.0"},
		{ID: 3, CVEID: "CVE-2023-3", PackageName: "pkg3", PackageVersion: "3.0"},
	}

	// Previous scan had: CVE-2 (persistent), CVE-4 (fixed)
	previousVulns := []models.Vulnerability{
		{ID: 2, CVEID: "CVE-2023-2", PackageName: "pkg2", PackageVersion: "2.0"},
		{ID: 4, CVEID: "CVE-2023-4", PackageName: "pkg4", PackageVersion: "4.0"},
	}

	mockScanRepo.On("GetByID", ctx, scanID).Return(currentScan, nil)
	mockScanRepo.On("GetPreviousScan", ctx, imageID, mock.Anything).Return(previousScan, nil)
	mockScanRepo.On("GetVulnerabilities", ctx, scanID).Return(currentVulns, nil)
	mockScanRepo.On("GetVulnerabilities", ctx, previousScanID).Return(previousVulns, nil)
	mockVulnRepo.On("MarkAsFixed", ctx, []int{4}).Return(nil)

	diff, err := analyzer.CompareScan(ctx, scanID)

	require.NoError(t, err)
	assert.Equal(t, scanID, diff.ScanID)
	assert.Equal(t, previousScanID, diff.PreviousScanID)

	// New: CVE-1, CVE-3
	assert.Len(t, diff.NewVulns, 2)
	// Fixed: CVE-4
	assert.Len(t, diff.FixedVulns, 1)
	assert.Equal(t, "CVE-2023-4", diff.FixedVulns[0].CVEID)
	// Persistent: CVE-2
	assert.Len(t, diff.PersistentVulns, 1)
	assert.Equal(t, "CVE-2023-2", diff.PersistentVulns[0].CVEID)

	assert.Equal(t, 2, diff.Summary.NewCount)
	assert.Equal(t, 1, diff.Summary.FixedCount)
	assert.Equal(t, 1, diff.Summary.PersistentCount)

	mockScanRepo.AssertExpectations(t)
	mockVulnRepo.AssertExpectations(t)
}
