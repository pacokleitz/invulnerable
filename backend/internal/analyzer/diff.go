package analyzer

import (
	"context"
	"fmt"

	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
)

// ScanRepository defines the interface for scan repository operations
type ScanRepository interface {
	GetByID(ctx context.Context, id int) (*models.Scan, error)
	GetPreviousScan(ctx context.Context, imageID int, currentScanDate string) (*models.Scan, error)
	GetVulnerabilities(ctx context.Context, scanID int) ([]models.Vulnerability, error)
}

// VulnerabilityRepository defines the interface for vulnerability repository operations
type VulnerabilityRepository interface {
	MarkAsFixed(ctx context.Context, vulnerabilityIDs []int) error
}

type Analyzer struct {
	scanRepo ScanRepository
	vulnRepo VulnerabilityRepository
}

func New(scanRepo ScanRepository, vulnRepo VulnerabilityRepository) *Analyzer {
	return &Analyzer{
		scanRepo: scanRepo,
		vulnRepo: vulnRepo,
	}
}

// NewWithRepositories creates an Analyzer with concrete db repositories
func NewWithRepositories(scanRepo *db.ScanRepository, vulnRepo *db.VulnerabilityRepository) *Analyzer {
	return &Analyzer{
		scanRepo: scanRepo,
		vulnRepo: vulnRepo,
	}
}

// CompareScan compares a scan with the previous scan for the same image
func (a *Analyzer) CompareScan(ctx context.Context, scanID int) (*models.ScanDiff, error) {
	return a.CompareScanWith(ctx, scanID, nil)
}

// CompareScanWith compares a scan with a specified previous scan, or the immediate previous scan if not specified
func (a *Analyzer) CompareScanWith(ctx context.Context, scanID int, previousScanID *int) (*models.ScanDiff, error) {
	// Get current scan
	currentScan, err := a.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current scan: %w", err)
	}

	// Get previous scan
	var previousScan *models.Scan
	if previousScanID != nil {
		// Use specified previous scan
		previousScan, err = a.scanRepo.GetByID(ctx, *previousScanID)
		if err != nil {
			return nil, fmt.Errorf("failed to get specified previous scan: %w", err)
		}
		// Verify it's for the same image
		if previousScan.ImageID != currentScan.ImageID {
			return nil, fmt.Errorf("previous scan is for a different image")
		}
	} else {
		// Get the immediate previous scan for the same image
		previousScan, err = a.scanRepo.GetPreviousScan(ctx, currentScan.ImageID, currentScan.ScanDate.Format("2006-01-02 15:04:05"))
		if err != nil {
			return nil, fmt.Errorf("failed to get previous scan: %w", err)
		}
	}

	// If no previous scan, all vulnerabilities are new
	if previousScan == nil {
		currentVulns, err := a.scanRepo.GetVulnerabilities(ctx, scanID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current vulnerabilities: %w", err)
		}

		return &models.ScanDiff{
			ScanID:          scanID,
			PreviousScanID:  0,
			NewVulns:        currentVulns,
			FixedVulns:      []models.Vulnerability{},
			PersistentVulns: []models.Vulnerability{},
			Summary: models.ScanDiffSummary{
				NewCount:        len(currentVulns),
				FixedCount:      0,
				PersistentCount: 0,
			},
		}, nil
	}

	// Get vulnerabilities from both scans
	currentVulns, err := a.scanRepo.GetVulnerabilities(ctx, scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current vulnerabilities: %w", err)
	}

	previousVulns, err := a.scanRepo.GetVulnerabilities(ctx, previousScan.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous vulnerabilities: %w", err)
	}

	// Create maps for efficient comparison
	currentMap := make(map[string]models.Vulnerability)
	for _, v := range currentVulns {
		key := makeVulnKey(v)
		currentMap[key] = v
	}

	previousMap := make(map[string]models.Vulnerability)
	for _, v := range previousVulns {
		key := makeVulnKey(v)
		previousMap[key] = v
	}

	// Categorize vulnerabilities
	newVulns := []models.Vulnerability{}
	fixedVulns := []models.Vulnerability{}
	persistentVulns := []models.Vulnerability{}

	// Find new and persistent vulnerabilities
	for key, vuln := range currentMap {
		if _, exists := previousMap[key]; exists {
			persistentVulns = append(persistentVulns, vuln)
		} else {
			newVulns = append(newVulns, vuln)
		}
	}

	// Find fixed vulnerabilities
	fixedVulnIDs := []int{}
	for key, vuln := range previousMap {
		if _, exists := currentMap[key]; !exists {
			fixedVulns = append(fixedVulns, vuln)
			fixedVulnIDs = append(fixedVulnIDs, vuln.ID)
		}
	}

	// Mark fixed vulnerabilities in database
	if len(fixedVulnIDs) > 0 {
		if err := a.vulnRepo.MarkAsFixed(ctx, fixedVulnIDs); err != nil {
			return nil, fmt.Errorf("failed to mark vulnerabilities as fixed: %w", err)
		}
	}

	return &models.ScanDiff{
		ScanID:          scanID,
		PreviousScanID:  previousScan.ID,
		NewVulns:        newVulns,
		FixedVulns:      fixedVulns,
		PersistentVulns: persistentVulns,
		Summary: models.ScanDiffSummary{
			NewCount:        len(newVulns),
			FixedCount:      len(fixedVulns),
			PersistentCount: len(persistentVulns),
		},
	}, nil
}

// makeVulnKey creates a unique key for vulnerability comparison
func makeVulnKey(v models.Vulnerability) string {
	return fmt.Sprintf("%s:%s:%s", v.CVEID, v.PackageName, v.PackageVersion)
}
