package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBuildSlackPayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	digest := "sha256:abc123"
	tests := []struct {
		name      string
		payload   NotificationPayload
		wantText  string
		wantColor string
	}{
		{
			name: "with vulnerabilities",
			payload: NotificationPayload{
				Image:      "nginx:latest",
				TotalVulns: 5,
				SeverityCounts: SeverityCounts{
					Critical: 1,
					High:     2,
					Medium:   1,
					Low:      1,
				},
				ScanID: 123,
			},
			wantText:  "⚠️ Found 5 vulnerabilities in `nginx:latest`",
			wantColor: "danger",
		},
		{
			name: "no vulnerabilities",
			payload: NotificationPayload{
				Image:      "alpine:latest",
				TotalVulns: 0,
				ScanID:     456,
			},
			wantText:  "✅ No vulnerabilities found in `alpine:latest`",
			wantColor: "good",
		},
		{
			name: "with image digest",
			payload: NotificationPayload{
				Image:       "postgres:15",
				ImageDigest: &digest,
				TotalVulns:  3,
				SeverityCounts: SeverityCounts{
					Medium: 3,
				},
				ScanID: 789,
			},
			wantText:  "⚠️ Found 3 vulnerabilities in `postgres:15`",
			wantColor: "#ffcc00",
		},
		{
			name: "with scan URL",
			payload: NotificationPayload{
				Image:      "redis:7",
				TotalVulns: 2,
				SeverityCounts: SeverityCounts{
					High: 2,
				},
				ScanID:  101,
				ScanURL: "http://localhost:3000/scans/101",
			},
			wantText:  "⚠️ Found 2 vulnerabilities in `redis:7`",
			wantColor: "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.buildSlackPayload(tt.payload)

			assert.Equal(t, tt.wantText, result.Text)
			assert.Len(t, result.Attachments, 1)
			assert.Equal(t, tt.wantColor, result.Attachments[0].Color)
			assert.Equal(t, "Vulnerability Summary", result.Attachments[0].Text)

			// Verify severity fields are present
			assert.GreaterOrEqual(t, len(result.Attachments[0].Fields), 4)

			// Check if digest field is present when expected
			if tt.payload.ImageDigest != nil {
				foundDigest := false
				for _, field := range result.Attachments[0].Fields {
					if field.Title == "Digest" {
						foundDigest = true
						assert.Equal(t, *tt.payload.ImageDigest, field.Value)
					}
				}
				assert.True(t, foundDigest, "Expected digest field to be present")
			}

			// Check if scan URL field is present when expected
			if tt.payload.ScanURL != "" {
				foundURL := false
				for _, field := range result.Attachments[0].Fields {
					if field.Title == "View Scan" {
						foundURL = true
						assert.Contains(t, field.Value, tt.payload.ScanURL)
					}
				}
				assert.True(t, foundURL, "Expected scan URL field to be present")
			}
		})
	}
}

func TestBuildSlackPayload_FieldValues(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	payload := NotificationPayload{
		Image:      "test:latest",
		TotalVulns: 10,
		SeverityCounts: SeverityCounts{
			Critical: 2,
			High:     3,
			Medium:   4,
			Low:      1,
		},
		ScanID: 1,
	}

	result := n.buildSlackPayload(payload)

	// Find and verify each severity field
	fields := result.Attachments[0].Fields
	for _, field := range fields {
		switch field.Title {
		case "Critical":
			assert.Equal(t, "2", field.Value)
			assert.True(t, field.Short)
		case "High":
			assert.Equal(t, "3", field.Value)
			assert.True(t, field.Short)
		case "Medium":
			assert.Equal(t, "4", field.Value)
			assert.True(t, field.Short)
		case "Low":
			assert.Equal(t, "1", field.Value)
			assert.True(t, field.Short)
		}
	}
}
