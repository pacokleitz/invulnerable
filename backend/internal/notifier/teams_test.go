package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBuildTeamsPayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	digest := "sha256:xyz789"
	tests := []struct {
		name       string
		payload    NotificationPayload
		wantTitle  string
		wantColor  string
		wantAction bool
	}{
		{
			name: "with vulnerabilities",
			payload: NotificationPayload{
				Image:      "nginx:latest",
				TotalVulns: 7,
				SeverityCounts: SeverityCounts{
					Critical: 2,
					High:     3,
					Medium:   1,
					Low:      1,
				},
				ScanID: 123,
			},
			wantTitle:  "Image Scan Results: nginx:latest",
			wantColor:  "FF0000",
			wantAction: false,
		},
		{
			name: "no vulnerabilities",
			payload: NotificationPayload{
				Image:      "alpine:latest",
				TotalVulns: 0,
				ScanID:     456,
			},
			wantTitle:  "✅ Image Scan Passed: alpine:latest",
			wantColor:  "00FF00",
			wantAction: false,
		},
		{
			name: "with image digest",
			payload: NotificationPayload{
				Image:       "postgres:15",
				ImageDigest: &digest,
				TotalVulns:  5,
				SeverityCounts: SeverityCounts{
					Medium: 5,
				},
				ScanID: 789,
			},
			wantTitle:  "Image Scan Results: postgres:15",
			wantColor:  "FFCC00",
			wantAction: false,
		},
		{
			name: "with scan URL",
			payload: NotificationPayload{
				Image:      "redis:7",
				TotalVulns: 4,
				SeverityCounts: SeverityCounts{
					High: 4,
				},
				ScanID:  101,
				ScanURL: "http://localhost:3000/scans/101",
			},
			wantTitle:  "Image Scan Results: redis:7",
			wantColor:  "FFA500",
			wantAction: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.buildTeamsPayload(tt.payload)

			assert.Equal(t, "MessageCard", result.Type)
			assert.Equal(t, "https://schema.org/extensions", result.Context)
			assert.Equal(t, tt.wantTitle, result.Title)
			assert.Equal(t, tt.wantColor, result.ThemeColor)
			assert.Len(t, result.Sections, 1)
			assert.Equal(t, "Vulnerability Summary", result.Sections[0].ActivityTitle)

			// Check for action button
			if tt.wantAction {
				assert.Len(t, result.PotentialAction, 1)
				assert.Equal(t, "OpenUri", result.PotentialAction[0].Type)
				assert.Equal(t, "View Scan Results", result.PotentialAction[0].Name)
				assert.Len(t, result.PotentialAction[0].Targets, 1)
				assert.Equal(t, tt.payload.ScanURL, result.PotentialAction[0].Targets[0].URI)
			} else {
				assert.Len(t, result.PotentialAction, 0)
			}

			// Verify facts contain severity counts
			facts := result.Sections[0].Facts
			assert.GreaterOrEqual(t, len(facts), 5) // 4 severities + total

			// Check if digest fact is present when expected
			if tt.payload.ImageDigest != nil {
				foundDigest := false
				for _, fact := range facts {
					if fact.Name == "Image Digest" {
						foundDigest = true
						assert.Equal(t, *tt.payload.ImageDigest, fact.Value)
					}
				}
				assert.True(t, foundDigest, "Expected digest fact to be present")
			}
		})
	}
}

func TestBuildTeamsPayload_FactValues(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	payload := NotificationPayload{
		Image:      "test:latest",
		TotalVulns: 15,
		SeverityCounts: SeverityCounts{
			Critical: 3,
			High:     5,
			Medium:   4,
			Low:      3,
		},
		ScanID: 1,
	}

	result := n.buildTeamsPayload(payload)

	// Find and verify each fact
	facts := result.Sections[0].Facts
	for _, fact := range facts {
		switch fact.Name {
		case "Critical":
			assert.Equal(t, "3", fact.Value)
		case "High":
			assert.Equal(t, "5", fact.Value)
		case "Medium":
			assert.Equal(t, "4", fact.Value)
		case "Low":
			assert.Equal(t, "3", fact.Value)
		case "Total Vulnerabilities":
			assert.Equal(t, "15", fact.Value)
		}
	}
}

func TestBuildTeamsPayload_Summary(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	tests := []struct {
		name        string
		totalVulns  int
		wantSummary string
	}{
		{
			name:        "with vulnerabilities",
			totalVulns:  10,
			wantSummary: "Found 10 vulnerabilities",
		},
		{
			name:        "no vulnerabilities",
			totalVulns:  0,
			wantSummary: "No vulnerabilities found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := NotificationPayload{
				Image:      "test:latest",
				TotalVulns: tt.totalVulns,
				ScanID:     1,
			}

			result := n.buildTeamsPayload(payload)
			assert.Equal(t, tt.wantSummary, result.Summary)
		})
	}
}
