package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/invulnerable/backend/internal/models"
)

type ImageRepository struct {
	db *Database
}

func NewImageRepository(db *Database) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(ctx context.Context, img *models.Image) error {
	query := `
		INSERT INTO images (registry, repository, tag, digest, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (registry, repository, tag)
		DO UPDATE SET digest = EXCLUDED.digest, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		img.Registry, img.Repository, img.Tag, img.Digest,
	).Scan(&img.ID, &img.CreatedAt, &img.UpdatedAt)
}

func (r *ImageRepository) GetByID(ctx context.Context, id int) (*models.Image, error) {
	var img models.Image
	query := `SELECT * FROM images WHERE id = $1`
	if err := r.db.GetContext(ctx, &img, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("image not found")
		}
		return nil, err
	}
	return &img, nil
}

func (r *ImageRepository) GetByName(ctx context.Context, registry, repository, tag string) (*models.Image, error) {
	var img models.Image
	query := `SELECT * FROM images WHERE registry = $1 AND repository = $2 AND tag = $3`
	if err := r.db.GetContext(ctx, &img, query, registry, repository, tag); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if not found
		}
		return nil, err
	}
	return &img, nil
}

func (r *ImageRepository) List(ctx context.Context, limit, offset int) ([]models.ImageWithStats, error) {
	query := `
		SELECT
			i.*,
			COUNT(DISTINCT s.id) as scan_count,
			MAX(s.scan_date) as last_scan_date,
			COUNT(DISTINCT CASE WHEN v.severity = 'Critical' AND v.status = 'active' THEN v.id END) as critical_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'High' AND v.status = 'active' THEN v.id END) as high_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Medium' AND v.status = 'active' THEN v.id END) as medium_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Low' AND v.status = 'active' THEN v.id END) as low_count
		FROM images i
		LEFT JOIN scans s ON s.image_id = i.id
		LEFT JOIN scan_vulnerabilities sv ON sv.scan_id = s.id
		LEFT JOIN vulnerabilities v ON v.id = sv.vulnerability_id
		GROUP BY i.id
		ORDER BY i.updated_at DESC
		LIMIT $1 OFFSET $2
	`
	images := []models.ImageWithStats{}
	if err := r.db.SelectContext(ctx, &images, query, limit, offset); err != nil {
		return nil, err
	}
	return images, nil
}

func (r *ImageRepository) GetScanHistory(ctx context.Context, imageID int, limit int) ([]models.ScanWithDetails, error) {
	query := `
		SELECT
			s.*,
			i.registry || '/' || i.repository || ':' || i.tag as image_name,
			COUNT(DISTINCT sv.vulnerability_id) as vulnerability_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Critical' THEN v.id END) as critical_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'High' THEN v.id END) as high_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Medium' THEN v.id END) as medium_count,
			COUNT(DISTINCT CASE WHEN v.severity = 'Low' THEN v.id END) as low_count
		FROM scans s
		JOIN images i ON i.id = s.image_id
		LEFT JOIN scan_vulnerabilities sv ON sv.scan_id = s.id
		LEFT JOIN vulnerabilities v ON v.id = sv.vulnerability_id
		WHERE s.image_id = $1
		GROUP BY s.id, i.registry, i.repository, i.tag
		ORDER BY s.scan_date DESC
		LIMIT $2
	`
	scans := []models.ScanWithDetails{}
	if err := r.db.SelectContext(ctx, &scans, query, imageID, limit); err != nil {
		return nil, err
	}
	return scans, nil
}
