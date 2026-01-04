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
	ScanOnlyFixable          bool      `db:"scan_only_fixable" json:"scan_only_fixable"`
	StatusChangeEnabled      bool      `db:"status_change_enabled" json:"status_change_enabled"`
	StatusChangeMinSeverity  string    `db:"status_change_min_severity" json:"status_change_min_severity"`
	StatusChangeOnlyFixable  bool      `db:"status_change_only_fixable" json:"status_change_only_fixable"`
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
	ScanOnlyFixable          bool     `json:"scan_only_fixable"`
	StatusChangeEnabled      bool     `json:"status_change_enabled"`
	StatusChangeMinSeverity  string   `json:"status_change_min_severity"`
	StatusChangeOnlyFixable  bool     `json:"status_change_only_fixable"`
	StatusChangeTransitions  []string `json:"status_change_transitions"`
	StatusChangeIncludeNotes bool     `json:"status_change_include_notes"`
}
