package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/invulnerable/backend/internal/models"
)

type SBOMRepository struct {
	db *Database
}

func NewSBOMRepository(db *Database) *SBOMRepository {
	return &SBOMRepository{db: db}
}

func (r *SBOMRepository) Create(ctx context.Context, sbom *models.SBOM) error {
	query := `
		INSERT INTO sboms (scan_id, format, version, document, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (scan_id)
		DO UPDATE SET format = EXCLUDED.format, version = EXCLUDED.version, document = EXCLUDED.document
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		sbom.ScanID, sbom.Format, sbom.Version, sbom.Document,
	).Scan(&sbom.ID, &sbom.CreatedAt)
}

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

func (r *SBOMRepository) GetDocumentByScanID(ctx context.Context, scanID int) (json.RawMessage, error) {
	var document json.RawMessage
	query := `SELECT document FROM sboms WHERE scan_id = $1`
	if err := r.db.QueryRowContext(ctx, query, scanID).Scan(&document); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("SBOM not found")
		}
		return nil, err
	}
	return document, nil
}
