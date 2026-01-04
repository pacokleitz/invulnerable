package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/invulnerable/backend/internal/models"
	"github.com/lib/pq"
)

// WebhookConfigRepository handles database operations for webhook configurations
type WebhookConfigRepository struct {
	db *Database
}

// NewWebhookConfigRepository creates a new webhook config repository
func NewWebhookConfigRepository(db *Database) *WebhookConfigRepository {
	return &WebhookConfigRepository{db: db}
}

// Upsert inserts or updates a webhook configuration
func (r *WebhookConfigRepository) Upsert(ctx context.Context, namespace, name string, req *models.WebhookConfigRequest) error {
	query := `
		INSERT INTO imagescan_webhook_configs (
			namespace, name,
			webhook_url, webhook_format,
			scan_min_severity, scan_only_fixable,
			status_change_enabled, status_change_min_severity, status_change_only_fixable,
			status_change_transitions, status_change_include_notes,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (namespace, name)
		DO UPDATE SET
			webhook_url = EXCLUDED.webhook_url,
			webhook_format = EXCLUDED.webhook_format,
			scan_min_severity = EXCLUDED.scan_min_severity,
			scan_only_fixable = EXCLUDED.scan_only_fixable,
			status_change_enabled = EXCLUDED.status_change_enabled,
			status_change_min_severity = EXCLUDED.status_change_min_severity,
			status_change_only_fixable = EXCLUDED.status_change_only_fixable,
			status_change_transitions = EXCLUDED.status_change_transitions,
			status_change_include_notes = EXCLUDED.status_change_include_notes,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		namespace, name,
		req.WebhookURL, req.WebhookFormat,
		req.ScanMinSeverity, req.ScanOnlyFixable,
		req.StatusChangeEnabled, req.StatusChangeMinSeverity, req.StatusChangeOnlyFixable,
		pq.Array(req.StatusChangeTransitions), req.StatusChangeIncludeNotes,
	)

	return err
}

// Get retrieves a webhook configuration by namespace and name
func (r *WebhookConfigRepository) Get(ctx context.Context, namespace, name string) (*models.WebhookConfig, error) {
	query := `
		SELECT id, namespace, name,
			webhook_url, webhook_format,
			scan_min_severity, scan_only_fixable,
			status_change_enabled, status_change_min_severity, status_change_only_fixable,
			status_change_transitions, status_change_include_notes,
			created_at, updated_at
		FROM imagescan_webhook_configs
		WHERE namespace = $1 AND name = $2
	`

	config := &models.WebhookConfig{}
	err := r.db.QueryRowContext(ctx, query, namespace, name).Scan(
		&config.ID, &config.Namespace, &config.Name,
		&config.WebhookURL, &config.WebhookFormat,
		&config.ScanMinSeverity, &config.ScanOnlyFixable,
		&config.StatusChangeEnabled, &config.StatusChangeMinSeverity, &config.StatusChangeOnlyFixable,
		pq.Array(&config.StatusChangeTransitions), &config.StatusChangeIncludeNotes,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook config: %w", err)
	}

	return config, nil
}

// Delete removes a webhook configuration
func (r *WebhookConfigRepository) Delete(ctx context.Context, namespace, name string) error {
	query := `DELETE FROM imagescan_webhook_configs WHERE namespace = $1 AND name = $2`
	_, err := r.db.ExecContext(ctx, query, namespace, name)
	return err
}
