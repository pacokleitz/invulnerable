package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/invulnerable/backend/internal/models"
	"github.com/invulnerable/backend/internal/storage"
)

type SBOMRepository struct {
	db      *Database
	storage storage.SBOMStorage
}

func NewSBOMRepository(db *Database, s3Storage storage.SBOMStorage) *SBOMRepository {
	return &SBOMRepository{
		db:      db,
		storage: s3Storage,
	}
}

// Create stores SBOM metadata in database and document in S3
func (r *SBOMRepository) Create(ctx context.Context, sbom *models.SBOM, document []byte) error {
	// Store document in S3 first
	if err := r.storage.Store(ctx, sbom.ScanID, document); err != nil {
		return fmt.Errorf("failed to store SBOM in S3: %w", err)
	}

	// Calculate size
	sizeBytes := int64(len(document))

	// Store metadata in database
	query := `
		INSERT INTO sboms (scan_id, format, version, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (scan_id)
		DO UPDATE SET format = EXCLUDED.format, version = EXCLUDED.version, size_bytes = EXCLUDED.size_bytes
		RETURNING id, created_at
	`
	if err := r.db.QueryRowContext(ctx, query,
		sbom.ScanID, sbom.Format, sbom.Version, sizeBytes,
	).Scan(&sbom.ID, &sbom.CreatedAt); err != nil {
		// Rollback: delete from S3 if database insert failed
		_ = r.storage.Delete(ctx, sbom.ScanID)
		return fmt.Errorf("failed to store SBOM metadata: %w", err)
	}

	sbom.SizeBytes = &sizeBytes
	return nil
}

// GetByScanID retrieves SBOM metadata from database
func (r *SBOMRepository) GetByScanID(ctx context.Context, scanID int) (*models.SBOM, error) {
	var sbom models.SBOM
	query := `SELECT * FROM sboms WHERE scan_id = $1`
	if err := r.db.GetContext(ctx, &sbom, query, scanID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("SBOM not found")
		}
		return nil, err
	}
	return &sbom, nil
}

// GetDocumentByScanID retrieves the SBOM document from S3
func (r *SBOMRepository) GetDocumentByScanID(ctx context.Context, scanID int) ([]byte, error) {
	// Verify SBOM exists in database
	if _, err := r.GetByScanID(ctx, scanID); err != nil {
		return nil, err
	}

	// Retrieve document from S3
	document, err := r.storage.Retrieve(ctx, scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve SBOM document from S3: %w", err)
	}

	return document, nil
}

// GetPresignedURL generates a pre-signed URL for direct SBOM download
func (r *SBOMRepository) GetPresignedURL(ctx context.Context, scanID int) (string, error) {
	// Verify SBOM exists in database
	if _, err := r.GetByScanID(ctx, scanID); err != nil {
		return "", err
	}

	// Generate presigned URL (valid for 1 hour)
	url, err := r.storage.GetPresignedURL(ctx, scanID, 3600)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// Delete removes SBOM from both S3 and database
func (r *SBOMRepository) Delete(ctx context.Context, scanID int) error {
	// Delete from S3 first
	if err := r.storage.Delete(ctx, scanID); err != nil {
		return fmt.Errorf("failed to delete SBOM from S3: %w", err)
	}

	// Delete metadata from database
	query := `DELETE FROM sboms WHERE scan_id = $1`
	if _, err := r.db.ExecContext(ctx, query, scanID); err != nil {
		return fmt.Errorf("failed to delete SBOM metadata: %w", err)
	}

	return nil
}
