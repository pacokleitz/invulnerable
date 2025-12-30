package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/invulnerable/backend/internal/models"
	"github.com/lib/pq"
)

type VulnerabilityRepository struct {
	db *Database
}

func NewVulnerabilityRepository(db *Database) *VulnerabilityRepository {
	return &VulnerabilityRepository{db: db}
}

func ValidateStatus(status string) error {
	for _, valid := range models.ValidStatuses {
		if status == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s (must be one of: %v)", status, models.ValidStatuses)
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

// ListWithImageInfo returns vulnerabilities with image context for compliance tracking
// Each row represents a unique vulnerability+image combination
func (r *VulnerabilityRepository) ListWithImageInfo(ctx context.Context, limit, offset int, severity, status *string, hasFix *bool, imageID *int, imageName, cveID *string) ([]models.VulnerabilityWithImageInfo, error) {
	// This query returns one row per image+vulnerability combination
	// showing when the vulnerability was first detected on that specific image
	query := `
		SELECT DISTINCT ON (v.id, i.id)
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
			v.last_seen_at,
			v.remediation_date,
			v.notes,
			v.created_at,
			v.updated_at,
			i.id as image_id,
			i.registry || '/' || i.repository || ':' || i.tag as image_name,
			i.digest as image_digest,
			MIN(s.scan_date) OVER (PARTITION BY v.id, i.id) as first_detected_at_for_image,
			FIRST_VALUE(s.id) OVER (PARTITION BY v.id, i.id ORDER BY s.scan_date DESC) as latest_scan_id,
			FIRST_VALUE(s.scan_date) OVER (PARTITION BY v.id, i.id ORDER BY s.scan_date DESC) as latest_scan_date,
			FIRST_VALUE(s.sla_critical) OVER (PARTITION BY v.id, i.id ORDER BY s.scan_date DESC) as sla_critical,
			FIRST_VALUE(s.sla_high) OVER (PARTITION BY v.id, i.id ORDER BY s.scan_date DESC) as sla_high,
			FIRST_VALUE(s.sla_medium) OVER (PARTITION BY v.id, i.id ORDER BY s.scan_date DESC) as sla_medium,
			FIRST_VALUE(s.sla_low) OVER (PARTITION BY v.id, i.id ORDER BY s.scan_date DESC) as sla_low
		FROM vulnerabilities v
		JOIN scan_vulnerabilities sv ON sv.vulnerability_id = v.id
		JOIN scans s ON s.id = sv.scan_id
		JOIN images i ON i.id = s.image_id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if severity != nil {
		query += fmt.Sprintf(" AND v.severity = $%d", argCount)
		args = append(args, *severity)
		argCount++
	}

	if status != nil {
		query += fmt.Sprintf(" AND v.status = $%d", argCount)
		args = append(args, *status)
		argCount++
	}

	if hasFix != nil {
		if *hasFix {
			query += " AND v.fix_version IS NOT NULL"
		} else {
			query += " AND v.fix_version IS NULL"
		}
	}

	if imageID != nil {
		query += fmt.Sprintf(" AND i.id = $%d", argCount)
		args = append(args, *imageID)
		argCount++
	}

	if imageName != nil {
		query += fmt.Sprintf(" AND (i.registry || '/' || i.repository || ':' || i.tag) ILIKE $%d", argCount)
		args = append(args, "%"+*imageName+"%")
		argCount++
	}

	if cveID != nil {
		query += fmt.Sprintf(" AND v.cve_id = $%d", argCount)
		args = append(args, *cveID)
		argCount++
	}

	query += ` ORDER BY
		v.id, i.id,
		CASE v.severity
			WHEN 'Critical' THEN 1
			WHEN 'High' THEN 2
			WHEN 'Medium' THEN 3
			WHEN 'Low' THEN 4
			ELSE 5
		END`

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	vulns := []models.VulnerabilityWithImageInfo{}
	if err := r.db.SelectContext(ctx, &vulns, query, args...); err != nil {
		return nil, err
	}
	return vulns, nil
}

func (r *VulnerabilityRepository) Update(ctx context.Context, id int, update *models.VulnerabilityUpdateWithContext) error {
	// Validate status if provided
	if update.Status != nil {
		if err := ValidateStatus(*update.Status); err != nil {
			return err
		}
	}

	// Get current state for audit trail
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get current vulnerability: %w", err)
	}

	// Build dynamic update query
	query := `UPDATE vulnerabilities SET updated_at = NOW()`
	args := []interface{}{}
	argCount := 1

	if update.Status != nil {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, *update.Status)
		argCount++

		// Auto-set remediation_date when marking as fixed
		if *update.Status == models.StatusFixed && current.RemediationDate == nil {
			query += fmt.Sprintf(", remediation_date = $%d", argCount)
			args = append(args, time.Now())
			argCount++
		}
	}

	if update.Notes != nil {
		query += fmt.Sprintf(", notes = $%d", argCount)
		args = append(args, *update.Notes)
		argCount++
	}

	if update.UpdatedBy != "" {
		query += fmt.Sprintf(", updated_by = $%d", argCount)
		args = append(args, update.UpdatedBy)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Create audit trail entries
	if update.Status != nil && current.Status != *update.Status {
		if err := r.CreateHistoryEntry(ctx, id, "status", &current.Status, update.Status, update.UpdatedBy, update.ImageID, update.ImageName); err != nil {
			// Log error but don't fail the update
			return fmt.Errorf("warning: failed to create status history: %w", err)
		}
	}

	if update.Notes != nil {
		oldNotes := ""
		if current.Notes != nil {
			oldNotes = *current.Notes
		}
		newNotes := *update.Notes
		if oldNotes != newNotes {
			if err := r.CreateHistoryEntry(ctx, id, "notes", &oldNotes, &newNotes, update.UpdatedBy, update.ImageID, update.ImageName); err != nil {
				return fmt.Errorf("warning: failed to create notes history: %w", err)
			}
		}
	}

	return nil
}

func (r *VulnerabilityRepository) BulkUpdate(ctx context.Context, ids []int, update *models.VulnerabilityUpdateWithContext) error {
	if len(ids) == 0 {
		return nil
	}

	// Get current state for all vulnerabilities (for audit trail)
	currentStates := make(map[int]*models.Vulnerability)
	for _, id := range ids {
		current, err := r.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get vulnerability %d: %w", id, err)
		}
		currentStates[id] = current
	}

	// Validate status if provided
	if update.Status != nil {
		if err := ValidateStatus(*update.Status); err != nil {
			return err
		}
	}

	// Build dynamic update query
	query := `UPDATE vulnerabilities SET updated_at = NOW()`
	args := []interface{}{}
	argCount := 1

	if update.Status != nil {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, *update.Status)
		argCount++

		// Auto-set remediation_date when marking as fixed (only if not already set)
		if *update.Status == models.StatusFixed {
			query += fmt.Sprintf(", remediation_date = COALESCE(remediation_date, $%d)", argCount)
			args = append(args, time.Now())
			argCount++
		}
	}

	if update.Notes != nil {
		query += fmt.Sprintf(", notes = $%d", argCount)
		args = append(args, *update.Notes)
		argCount++
	}

	if update.UpdatedBy != "" {
		query += fmt.Sprintf(", updated_by = $%d", argCount)
		args = append(args, update.UpdatedBy)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = ANY($%d)", argCount)
	args = append(args, pq.Array(ids))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Create audit trail entries for each vulnerability
	for _, id := range ids {
		current := currentStates[id]

		if update.Status != nil && current.Status != *update.Status {
			_ = r.CreateHistoryEntry(ctx, id, "status", &current.Status, update.Status, update.UpdatedBy, update.ImageID, update.ImageName)
			// Ignore error - audit trail is best-effort
		}

		if update.Notes != nil {
			oldNotes := ""
			if current.Notes != nil {
				oldNotes = *current.Notes
			}
			newNotes := *update.Notes
			if oldNotes != newNotes {
				_ = r.CreateHistoryEntry(ctx, id, "notes", &oldNotes, &newNotes, update.UpdatedBy, update.ImageID, update.ImageName)
				// Ignore error - audit trail is best-effort
			}
		}
	}

	return nil
}

func (r *VulnerabilityRepository) MarkAsFixed(ctx context.Context, vulnerabilityIDs []int) error {
	if len(vulnerabilityIDs) == 0 {
		return nil
	}

	// Get current state for audit trail
	currentStates := make(map[int]string)
	for _, id := range vulnerabilityIDs {
		current, err := r.GetByID(ctx, id)
		if err == nil && current.Status != models.StatusFixed {
			currentStates[id] = current.Status
		}
	}

	query := `
		UPDATE vulnerabilities
		SET status = 'fixed', remediation_date = $1, updated_at = NOW(), updated_by = 'system'
		WHERE id = ANY($2) AND status != 'fixed'
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), pq.Array(vulnerabilityIDs))
	if err != nil {
		return err
	}

	// Create audit entries for automatic fixes
	for id, oldStatus := range currentStates {
		newStatus := models.StatusFixed
		_ = r.CreateHistoryEntry(ctx, id, "status", &oldStatus, &newStatus, "system", nil, nil)
		// Ignore error - audit trail is best-effort
	}

	return nil
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
			return nil, fmt.Errorf("vulnerability not found")
		}
		return nil, err
	}
	return &vuln, nil
}

func (r *VulnerabilityRepository) CreateHistoryEntry(ctx context.Context, vulnerabilityID int, fieldName string, oldValue, newValue *string, changedBy string, imageID *int, imageName *string) error {
	query := `
		INSERT INTO vulnerability_history (
			vulnerability_id, field_name, old_value, new_value,
			changed_by, changed_at, image_id, image_name
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		vulnerabilityID, fieldName, oldValue, newValue,
		changedBy, imageID, imageName)
	return err
}

func (r *VulnerabilityRepository) GetHistory(ctx context.Context, vulnerabilityID int) ([]models.VulnerabilityHistory, error) {
	history := []models.VulnerabilityHistory{}
	query := `
		SELECT * FROM vulnerability_history
		WHERE vulnerability_id = $1
		ORDER BY changed_at DESC
	`
	err := r.db.SelectContext(ctx, &history, query, vulnerabilityID)
	if err != nil {
		return nil, err
	}
	// Ensure we always return an empty slice, never nil
	if history == nil {
		history = []models.VulnerabilityHistory{}
	}
	return history, nil
}
