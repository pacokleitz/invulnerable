package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/invulnerable/backend/internal/models"
)

type ScanRepository struct {
	db *Database
}

func NewScanRepository(db *Database) *ScanRepository {
	return &ScanRepository{db: db}
}

func (r *ScanRepository) Create(ctx context.Context, scan *models.Scan) error {
	query := `
		INSERT INTO scans (image_id, scan_date, syft_version, grype_version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		scan.ImageID, scan.ScanDate, scan.SyftVersion, scan.GrypeVersion, scan.Status,
	).Scan(&scan.ID, &scan.CreatedAt, &scan.UpdatedAt)
}

func (r *ScanRepository) GetByID(ctx context.Context, id int) (*models.Scan, error) {
	var scan models.Scan
	query := `SELECT * FROM scans WHERE id = $1`
	if err := r.db.GetContext(ctx, &scan, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scan not found")
		}
		return nil, err
	}
	return &scan, nil
}

func (r *ScanRepository) GetWithDetails(ctx context.Context, id int, hasFix *bool) (*models.ScanWithDetails, error) {
	// Build fix filter
	fixFilter := "1=1"
	if hasFix != nil {
		if *hasFix {
			fixFilter = "v.fix_version IS NOT NULL"
		} else {
			fixFilter = "v.fix_version IS NULL"
		}
	}

	query := `
		SELECT
			s.*,
			i.registry || '/' || i.repository || ':' || i.tag as image_name,
			i.digest as image_digest,
			COUNT(DISTINCT CASE WHEN ` + fixFilter + ` THEN sv.vulnerability_id END) as vulnerability_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Critical' AND ` + fixFilter + ` THEN v.id END) as critical_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'High' AND ` + fixFilter + ` THEN v.id END) as high_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Medium' AND ` + fixFilter + ` THEN v.id END) as medium_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Low' AND ` + fixFilter + ` THEN v.id END) as low_count
		FROM scans s
		JOIN images i ON i.id = s.image_id
		LEFT JOIN scan_vulnerabilities sv ON sv.scan_id = s.id
		LEFT JOIN vulnerabilities v ON v.id = sv.vulnerability_id
		WHERE s.id = $1
		GROUP BY s.id, i.registry, i.repository, i.tag, i.digest
	`
	var scan models.ScanWithDetails
	if err := r.db.GetContext(ctx, &scan, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scan not found")
		}
		return nil, err
	}
	return &scan, nil
}

func (r *ScanRepository) List(ctx context.Context, limit, offset int, imageID *int, hasFix *bool) ([]models.ScanWithDetails, error) {
	// Build fix filter
	fixFilter := "1=1"
	if hasFix != nil {
		if *hasFix {
			fixFilter = "v.fix_version IS NOT NULL"
		} else {
			fixFilter = "v.fix_version IS NULL"
		}
	}

	query := `
		SELECT
			s.*,
			i.registry || '/' || i.repository || ':' || i.tag as image_name,
			i.digest as image_digest,
			COUNT(DISTINCT CASE WHEN ` + fixFilter + ` THEN sv.vulnerability_id END) as vulnerability_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Critical' AND ` + fixFilter + ` THEN v.id END) as critical_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'High' AND ` + fixFilter + ` THEN v.id END) as high_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Medium' AND ` + fixFilter + ` THEN v.id END) as medium_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Low' AND ` + fixFilter + ` THEN v.id END) as low_count
		FROM scans s
		JOIN images i ON i.id = s.image_id
		LEFT JOIN scan_vulnerabilities sv ON sv.scan_id = s.id
		LEFT JOIN vulnerabilities v ON v.id = sv.vulnerability_id
	`
	args := []interface{}{}
	if imageID != nil {
		query += ` WHERE s.image_id = $1`
		args = append(args, *imageID)
		query += ` GROUP BY s.id, i.registry, i.repository, i.tag, i.digest ORDER BY s.scan_date DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` GROUP BY s.id, i.registry, i.repository, i.tag, i.digest ORDER BY s.scan_date DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}

	scans := []models.ScanWithDetails{}
	if err := r.db.SelectContext(ctx, &scans, query, args...); err != nil {
		return nil, err
	}
	return scans, nil
}

func (r *ScanRepository) GetPreviousScan(ctx context.Context, imageID int, currentScanDate string) (*models.Scan, error) {
	var scan models.Scan
	query := `
		SELECT * FROM scans
		WHERE image_id = $1 AND scan_date < $2
		ORDER BY scan_date DESC
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &scan, query, imageID, currentScanDate); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No previous scan
		}
		return nil, err
	}
	return &scan, nil
}

func (r *ScanRepository) GetVulnerabilities(ctx context.Context, scanID int) ([]models.Vulnerability, error) {
	query := `
		SELECT
			v.id,
			v.cve_id,
			v.package_name,
			v.package_version,
			v.package_type,
			v.severity,
			v.fix_version,
			v.url,
			v.description,
			v.status,
			-- Calculate first detection for this specific image
			(
				SELECT MIN(s2.scan_date)
				FROM scan_vulnerabilities sv2
				JOIN scans s2 ON s2.id = sv2.scan_id
				WHERE sv2.vulnerability_id = v.id
					AND s2.image_id = (SELECT image_id FROM scans WHERE id = $1)
			) as first_detected_at,
			v.last_seen_at,
			v.remediation_date,
			v.notes,
			v.created_at,
			v.updated_at
		FROM vulnerabilities v
		JOIN scan_vulnerabilities sv ON sv.vulnerability_id = v.id
		WHERE sv.scan_id = $1
		ORDER BY
			CASE v.severity
				WHEN 'Critical' THEN 1
				WHEN 'High' THEN 2
				WHEN 'Medium' THEN 3
				WHEN 'Low' THEN 4
				ELSE 5
			END,
			v.cve_id
	`
	vulns := []models.Vulnerability{}
	if err := r.db.SelectContext(ctx, &vulns, query, scanID); err != nil {
		return nil, err
	}
	return vulns, nil
}
