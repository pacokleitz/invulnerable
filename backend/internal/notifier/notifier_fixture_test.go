package notifier

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// loadGrypeFixture loads a Grype output fixture from testdata
func loadGrypeFixture(t *testing.T, filename string) models.GrypeResult {
	t.Helper()

	fixturePath := filepath.Join("..", "testdata", filename)
	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "failed to read fixture file")

	var result models.GrypeResult
	err = json.Unmarshal(data, &result)
	require.NoError(t, err, "failed to unmarshal fixture")

	return result
}

func TestWebhookNotification_OnlyFixable_Mixed(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-mixed.json")

	// Mock webhook server
	webhookCalled := false
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	notifier := New(logger, "http://example.com")

	// Test with onlyFixable=true
	config := WebhookConfig{
		URL:         server.URL,
		Format:      "slack",
		MinSeverity: "Low", // Include all severities
		OnlyFixable: true,  // Only notify for fixed CVEs
	}

	// Filter matches like the scan handler does
	filteredMatches := filterByOnlyFixable(grypeResult.Matches, config.OnlyFixable)

	// Build payload with filtered matches
	severityCounts := SeverityCounts{}
	for _, match := range filteredMatches {
		switch match.Vulnerability.Severity {
		case "Critical":
			severityCounts.Critical++
		case "High":
			severityCounts.High++
		case "Medium":
			severityCounts.Medium++
		case "Low":
			severityCounts.Low++
		default:
			severityCounts.Negligible++
		}
	}

	payload := NotificationPayload{
		Image:          "nginx:latest",
		ScanID:         1,
		TotalVulns:     len(filteredMatches),
		SeverityCounts: severityCounts,
		ScanURL:        "http://example.com/scans/1",
	}

	err := notifier.SendNotification(context.Background(), config, payload)
	require.NoError(t, err)

	assert.True(t, webhookCalled, "webhook should be called")
	assert.Equal(t, 2, payload.TotalVulns, "should only count fixed CVEs")
	assert.Equal(t, 1, payload.SeverityCounts.Critical, "should have 1 Critical (fixed)")
	assert.Equal(t, 0, payload.SeverityCounts.High, "should have 0 High (unfixed)")
	assert.Equal(t, 1, payload.SeverityCounts.Medium, "should have 1 Medium (fixed)")
	assert.Equal(t, 0, payload.SeverityCounts.Low, "should have 0 Low (unfixed)")
}

func TestWebhookNotification_OnlyFixable_AllUnfixed(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-all-unfixed.json")

	// Mock webhook server
	webhookCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	notifier := New(logger, "http://example.com")

	// Test with onlyFixable=true
	config := WebhookConfig{
		URL:         server.URL,
		Format:      "slack",
		MinSeverity: "Low",
		OnlyFixable: true, // Only notify for fixed CVEs
	}

	// Filter matches
	filteredMatches := filterByOnlyFixable(grypeResult.Matches, config.OnlyFixable)

	payload := NotificationPayload{
		Image:      "test:unfixed",
		ScanID:     1,
		TotalVulns: len(filteredMatches),
		ScanURL:    "http://example.com/scans/1",
	}

	err := notifier.SendNotification(context.Background(), config, payload)
	require.NoError(t, err)

	// Webhook should not be called because TotalVulns=0 (no fixed CVEs)
	assert.False(t, webhookCalled, "webhook should not be called when no fixed CVEs")
	assert.Equal(t, 0, payload.TotalVulns, "should have 0 vulnerabilities (all unfixed)")
}

func TestWebhookNotification_OnlyFixable_Disabled(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-all-unfixed.json")

	// Mock webhook server
	webhookCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	notifier := New(logger, "http://example.com")

	// Test with onlyFixable=false (disabled)
	config := WebhookConfig{
		URL:         server.URL,
		Format:      "slack",
		MinSeverity: "Low",
		OnlyFixable: false, // Include all CVEs
	}

	// Don't filter matches
	filteredMatches := filterByOnlyFixable(grypeResult.Matches, config.OnlyFixable)

	// Count severities
	severityCounts := SeverityCounts{}
	for _, match := range filteredMatches {
		switch match.Vulnerability.Severity {
		case "Critical":
			severityCounts.Critical++
		case "High":
			severityCounts.High++
		}
	}

	payload := NotificationPayload{
		Image:          "test:unfixed",
		ScanID:         1,
		TotalVulns:     len(filteredMatches),
		SeverityCounts: severityCounts,
		ScanURL:        "http://example.com/scans/1",
	}

	err := notifier.SendNotification(context.Background(), config, payload)
	require.NoError(t, err)

	assert.True(t, webhookCalled, "webhook should be called when onlyFixable=false")
	assert.Equal(t, 2, payload.TotalVulns, "should include all CVEs (even unfixed)")
}

func TestStatusChangeNotification_OnlyFixable(t *testing.T) {
	logger := zap.NewNop()
	notifier := New(logger, "http://example.com")

	// Mock webhook server
	webhookCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name          string
		onlyFixable   bool
		hasFixVersion bool
		expectWebhook bool
	}{
		{
			name:          "OnlyFixable=true with fix - should notify",
			onlyFixable:   true,
			hasFixVersion: true,
			expectWebhook: true,
		},
		{
			name:          "OnlyFixable=true without fix - should not notify",
			onlyFixable:   true,
			hasFixVersion: false,
			expectWebhook: false,
		},
		{
			name:          "OnlyFixable=false without fix - should notify",
			onlyFixable:   false,
			hasFixVersion: false,
			expectWebhook: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhookCalled = false

			config := StatusChangeWebhookConfig{
				URL:         server.URL,
				Format:      "slack",
				MinSeverity: "Low",
				OnlyFixable: tt.onlyFixable,
			}

			var fixVersion *string
			if tt.hasFixVersion {
				fix := "1.2.3"
				fixVersion = &fix
			}

			payload := StatusChangeNotificationPayload{
				CVEID:           "CVE-2024-TEST",
				PackageName:     "testpkg",
				PackageVersion:  "1.0.0",
				Severity:        "High",
				FixVersion:      fixVersion,
				OldStatus:       "active",
				NewStatus:       "in_progress",
				ChangedBy:       "test@example.com",
				ImageName:       "test:latest",
				VulnerabilityID: 1,
			}

			err := notifier.SendStatusChangeNotification(context.Background(), config, payload)
			require.NoError(t, err)

			if tt.expectWebhook {
				assert.True(t, webhookCalled, "webhook should be called")
			} else {
				assert.False(t, webhookCalled, "webhook should not be called")
			}
		})
	}
}

// Helper function
func filterByOnlyFixable(matches []models.GrypeMatch, onlyFixable bool) []models.GrypeMatch {
	if !onlyFixable {
		return matches
	}

	filtered := []models.GrypeMatch{}
	for _, match := range matches {
		if len(match.Vulnerability.Fix.Versions) > 0 {
			filtered = append(filtered, match)
		}
	}
	return filtered
}
