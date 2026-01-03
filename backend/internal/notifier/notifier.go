package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Notifier struct {
	logger      *zap.Logger
	httpClient  *http.Client
	frontendURL string
}

func New(logger *zap.Logger, frontendURL string) *Notifier {
	return &Notifier{
		logger:      logger,
		frontendURL: frontendURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// WebhookConfig represents webhook configuration from scan request
type WebhookConfig struct {
	URL         string `json:"url"`
	Format      string `json:"format"`
	MinSeverity string `json:"min_severity"`
	OnlyFixed   bool   `json:"only_fixed"`
}

// NotificationPayload contains data for webhook notification
type NotificationPayload struct {
	Image           string
	ImageDigest     *string
	ScanID          int
	TotalVulns      int
	SeverityCounts  SeverityCounts
	VulnsBySeverity map[string][]VulnerabilityInfo
	ScanURL         string
}

type SeverityCounts struct {
	Critical   int
	High       int
	Medium     int
	Low        int
	Negligible int
}

type VulnerabilityInfo struct {
	CVEID       string
	PackageName string
	Severity    string
	FixVersion  *string
	HasFix      bool
}

// SendNotification sends webhook notification
func (n *Notifier) SendNotification(ctx context.Context, config WebhookConfig, payload NotificationPayload) error {
	// Check if there are any vulnerabilities to notify about (e.g., when onlyFixed filters everything out)
	if payload.TotalVulns == 0 {
		n.logger.Info("no vulnerabilities to notify about, skipping notification",
			zap.Bool("only_fixed", config.OnlyFixed),
			zap.Int("scan_id", payload.ScanID))
		return nil
	}

	// Check if notification should be sent based on severity threshold
	if !n.shouldNotify(config.MinSeverity, payload.SeverityCounts) {
		n.logger.Info("no vulnerabilities meet severity threshold, skipping notification",
			zap.String("min_severity", config.MinSeverity),
			zap.Int("scan_id", payload.ScanID))
		return nil
	}

	// Construct scan URL if frontend URL is configured
	if n.frontendURL != "" && payload.ScanURL == "" {
		payload.ScanURL = fmt.Sprintf("%s/scans/%d", n.frontendURL, payload.ScanID)
	}

	var webhookPayload interface{}

	switch config.Format {
	case "teams":
		webhookPayload = n.buildTeamsPayload(payload)
	default:
		// Default to Slack format for backward compatibility
		webhookPayload = n.buildSlackPayload(payload)
	}

	return n.sendWebhook(ctx, config.URL, webhookPayload)
}

// shouldNotify determines if notification should be sent based on severity threshold
func (n *Notifier) shouldNotify(minSeverity string, counts SeverityCounts) bool {
	severityOrder := map[string]int{
		"Critical":   5,
		"High":       4,
		"Medium":     3,
		"Low":        2,
		"Negligible": 1,
	}

	threshold, ok := severityOrder[minSeverity]
	if !ok {
		// Unknown severity, default to High
		threshold = 4
	}

	// Check if any vulnerabilities meet or exceed threshold
	if threshold <= 5 && counts.Critical > 0 {
		return true
	}
	if threshold <= 4 && counts.High > 0 {
		return true
	}
	if threshold <= 3 && counts.Medium > 0 {
		return true
	}
	if threshold <= 2 && counts.Low > 0 {
		return true
	}
	if threshold <= 1 && counts.Negligible > 0 {
		return true
	}

	return false
}

func (n *Notifier) sendWebhook(ctx context.Context, url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	n.logger.Info("webhook notification sent successfully",
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// StatusChangeNotificationPayload contains data for status change webhooks
type StatusChangeNotificationPayload struct {
	CVEID           string
	PackageName     string
	PackageVersion  string
	Severity        string
	FixVersion      *string
	OldStatus       string
	NewStatus       string
	ChangedBy       string
	Notes           *string
	ImageName       string
	VulnerabilityID int
	VulnURL         string
	Timestamp       time.Time
}

// StatusChangeWebhookConfig extends webhook config for status changes
type StatusChangeWebhookConfig struct {
	URL                string
	Format             string
	MinSeverity        string
	OnlyFixed          bool
	StatusTransitions  []string
	IncludeNoteChanges bool
}

// SendStatusChangeNotification sends webhook for vulnerability status changes
func (n *Notifier) SendStatusChangeNotification(ctx context.Context, config StatusChangeWebhookConfig, payload StatusChangeNotificationPayload) error {
	// Check severity threshold
	if !n.shouldNotifyStatusChange(config.MinSeverity, payload.Severity) {
		n.logger.Info("vulnerability does not meet severity threshold",
			zap.String("severity", payload.Severity),
			zap.String("min_severity", config.MinSeverity))
		return nil
	}

	// Check if only fixed CVEs should trigger notifications
	if config.OnlyFixed && payload.FixVersion == nil {
		n.logger.Info("vulnerability has no fix available, skipping notification",
			zap.String("cve_id", payload.CVEID),
			zap.Bool("only_fixed", config.OnlyFixed))
		return nil
	}

	// Check status transition filter
	if len(config.StatusTransitions) > 0 {
		transition := fmt.Sprintf("%s→%s", payload.OldStatus, payload.NewStatus)
		if !contains(config.StatusTransitions, transition) {
			n.logger.Info("status transition not in filter",
				zap.String("transition", transition))
			return nil
		}
	}

	// Build URL
	if n.frontendURL != "" && payload.VulnURL == "" {
		payload.VulnURL = fmt.Sprintf("%s/vulnerabilities/%d", n.frontendURL, payload.VulnerabilityID)
	}

	var webhookPayload interface{}
	switch config.Format {
	case "teams":
		webhookPayload = n.buildTeamsStatusChangePayload(payload)
	default:
		webhookPayload = n.buildSlackStatusChangePayload(payload)
	}

	return n.sendWebhook(ctx, config.URL, webhookPayload)
}

func (n *Notifier) shouldNotifyStatusChange(minSeverity, vulnSeverity string) bool {
	severityOrder := map[string]int{
		"Critical":   5,
		"High":       4,
		"Medium":     3,
		"Low":        2,
		"Negligible": 1,
	}

	threshold := severityOrder[minSeverity]
	vulnLevel := severityOrder[vulnSeverity]

	return vulnLevel >= threshold
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
