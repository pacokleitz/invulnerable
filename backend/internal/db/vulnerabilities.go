package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/invulnerable/backend/internal/models"
)

type VulnerabilityRepository struct {
	db *Database
}

func NewVulnerabilityRepository(db *Database) *VulnerabilityRepository {
	return &VulnerabilityRepository{db: db}
}

func (r *VulnerabilityRepository) Upsert(ctx context.Context, vuln *models.Vulnerability) error {
	query := `
		INSERT INTO vulnerabilities (
			cve_id, package_name, package_version, package_type,
			severity, fix_version, url, description, status,
			first_detected_at, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (cve_id, package_name, package_version)
		DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			severity = EXCLUDED.severity,
			fix_version = EXCLUDED.fix_version,
			url = EXCLUDED.url,
			description = EXCLUDED.description,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		vuln.CVEID, vuln.PackageName, vuln.PackageVersion, vuln.PackageType,
		vuln.Severity, vuln.FixVersion, vuln.URL, vuln.Description, vuln.Status,
		vuln.FirstDetectedAt, vuln.LastSeenAt,
	).Scan(&vuln.ID, &vuln.CreatedAt, &vuln.UpdatedAt)
}

func (r *VulnerabilityRepository) GetByID(ctx context.Context, id int) (*models.Vulnerability, error) {
	var vuln models.Vulnerability
	query := `SELECT * FROM vulnerabilities WHERE id = $1`
	if err := r.db.GetContext(ctx, &vuln, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vulnerability not found")
		}
		return nil, err
	}
	return &vuln, nil
}

func (r *VulnerabilityRepository) GetByCVE(ctx context.Context, cveID string) ([]models.Vulnerability, error) {
	vulns := []models.Vulnerability{}
	query := `SELECT * FROM vulnerabilities WHERE cve_id = $1 ORDER BY package_name, package_version`
	if err := r.db.SelectContext(ctx, &vulns, query, cveID); err != nil {
		return nil, err
	}
	return vulns, nil
}

func (r *VulnerabilityRepository) List(ctx context.Context, limit, offset int, severity, status *string, hasFix *bool) ([]models.Vulnerability, error) {
	query := `SELECT * FROM vulnerabilities WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, *severity)
		argCount++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *status)
		argCount++
	}

	// Filter by fix availability
	if hasFix != nil {
		if *hasFix {
			query += " AND fix_version IS NOT NULL"
		} else {
			query += " AND fix_version IS NULL"
		}
	}

	query += ` ORDER BY
		CASE severity
			WHEN 'Critical' THEN 1
			WHEN 'High' THEN 2
			WHEN 'Medium' THEN 3
			WHEN 'Low' THEN 4
			ELSE 5
		END,
		first_detected_at DESC`

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	vulns := []models.Vulnerability{}
	if err := r.db.SelectContext(ctx, &vulns, query, args...); err != nil {
		return nil, err
	}
	return vulns, nil
}

func (r *VulnerabilityRepository) Update(ctx context.Context, id int, update *models.VulnerabilityUpdate) error {
	query := `UPDATE vulnerabilities SET updated_at = NOW()`
	args := []interface{}{}
	argCount := 1

	if update.Status != nil {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, *update.Status)
		argCount++
	}

	if update.Notes != nil {
		query += fmt.Sprintf(", notes = $%d", argCount)
		args = append(args, *update.Notes)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *VulnerabilityRepository) MarkAsFixed(ctx context.Context, vulnerabilityIDs []int) error {
	if len(vulnerabilityIDs) == 0 {
		return nil
	}

	query := `
		UPDATE vulnerabilities
		SET status = 'fixed', remediation_date = $1, updated_at = NOW()
		WHERE id = ANY($2) AND status = 'active'
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), vulnerabilityIDs)
	return err
}

func (r *VulnerabilityRepository) LinkToScan(ctx context.Context, scanID, vulnerabilityID int) error {
	query := `
		INSERT INTO scan_vulnerabilities (scan_id, vulnerability_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (scan_id, vulnerability_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, scanID, vulnerabilityID)
	return err
}

func (r *VulnerabilityRepository) GetByUniqueKey(ctx context.Context, cveID, packageName, packageVersion string) (*models.Vulnerability, error) {
	var vuln models.Vulnerability
	query := `SELECT * FROM vulnerabilities WHERE cve_id = $1 AND package_name = $2 AND package_version = $3`
	if err := r.db.GetContext(ctx, &vuln, query, cveID, packageName, packageVersion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &vuln, nil
}
