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
}

// SendNotification sends webhook notification
func (n *Notifier) SendNotification(ctx context.Context, config WebhookConfig, payload NotificationPayload) error {
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
