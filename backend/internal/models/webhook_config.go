package models

import "time"

// WebhookConfig represents a webhook configuration synced from an ImageScan CRD
type WebhookConfig struct {
	ID                       int       `db:"id" json:"id"`
	Namespace                string    `db:"namespace" json:"namespace"`
	Name                     string    `db:"name" json:"name"`
	WebhookURL               string    `db:"webhook_url" json:"webhook_url"`
	WebhookFormat            string    `db:"webhook_format" json:"webhook_format"`
	ScanMinSeverity          string    `db:"scan_min_severity" json:"scan_min_severity"`
	ScanOnlyFixed            bool      `db:"scan_only_fixed" json:"scan_only_fixed"`
	StatusChangeEnabled      bool      `db:"status_change_enabled" json:"status_change_enabled"`
	StatusChangeMinSeverity  string    `db:"status_change_min_severity" json:"status_change_min_severity"`
	StatusChangeOnlyFixed    bool      `db:"status_change_only_fixed" json:"status_change_only_fixed"`
	StatusChangeTransitions  []string  `db:"status_change_transitions" json:"status_change_transitions"`
	StatusChangeIncludeNotes bool      `db:"status_change_include_notes" json:"status_change_include_notes"`
	CreatedAt                time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time `db:"updated_at" json:"updated_at"`
}

// WebhookConfigRequest is the API request format for upserting webhook configs
type WebhookConfigRequest struct {
	WebhookURL               string   `json:"webhook_url"`
	WebhookFormat            string   `json:"webhook_format"`
	ScanMinSeverity          string   `json:"scan_min_severity"`
	ScanOnlyFixed            bool     `json:"scan_only_fixed"`
	StatusChangeEnabled      bool     `json:"status_change_enabled"`
	StatusChangeMinSeverity  string   `json:"status_change_min_severity"`
	StatusChangeOnlyFixed    bool     `json:"status_change_only_fixed"`
	StatusChangeTransitions  []string `json:"status_change_transitions"`
	StatusChangeIncludeNotes bool     `json:"status_change_include_notes"`
}
