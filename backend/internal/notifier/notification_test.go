package notifier

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSendNotification_Success(t *testing.T) {
	// Create mock webhook server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify payload can be decoded
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	config := WebhookConfig{
		URL:         server.URL,
		Format:      "slack",
		MinSeverity: "Low",
	}

	payload := NotificationPayload{
		Image:      "nginx:latest",
		TotalVulns: 5,
		SeverityCounts: SeverityCounts{
			Critical: 1,
			High:     2,
			Medium:   1,
			Low:      1,
		},
		ScanID: 123,
	}

	err := n.SendNotification(context.Background(), config, payload)
	assert.NoError(t, err)
}

func TestSendNotification_TeamsFormat(t *testing.T) {
	// Create mock webhook server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var payload TeamsPayload
		err := json.Unmarshal(body, &payload)
		require.NoError(t, err)

		assert.Equal(t, "MessageCard", payload.Type)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	config := WebhookConfig{
		URL:         server.URL,
		Format:      "teams",
		MinSeverity: "Medium",
	}

	payload := NotificationPayload{
		Image:      "postgres:15",
		TotalVulns: 3,
		SeverityCounts: SeverityCounts{
			Medium: 3,
		},
		ScanID: 456,
	}

	err := n.SendNotification(context.Background(), config, payload)
	assert.NoError(t, err)
}

func TestSendNotification_NoVulnerabilities(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	config := WebhookConfig{
		URL:         "http://example.com/webhook",
		MinSeverity: "Low",
	}

	payload := NotificationPayload{
		Image:      "alpine:latest",
		TotalVulns: 0,
		ScanID:     789,
	}

	// Should not send notification when there are no vulnerabilities
	err := n.SendNotification(context.Background(), config, payload)
	assert.NoError(t, err)
}

func TestSendNotification_BelowSeverityThreshold(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	config := WebhookConfig{
		URL:         "http://example.com/webhook",
		MinSeverity: "Critical",
	}

	payload := NotificationPayload{
		Image:      "redis:7",
		TotalVulns: 5,
		SeverityCounts: SeverityCounts{
			High:   3,
			Medium: 2,
		},
		ScanID: 101,
	}

	// Should not send notification when no vulnerabilities meet threshold
	err := n.SendNotification(context.Background(), config, payload)
	assert.NoError(t, err)
}

func TestSendNotification_ScanURLInjection(t *testing.T) {
	// Create mock webhook server
	receivedPayload := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedPayload <- string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://frontend.local")

	config := WebhookConfig{
		URL:         server.URL,
		Format:      "slack",
		MinSeverity: "Low",
	}

	payload := NotificationPayload{
		Image:      "test:latest",
		TotalVulns: 1,
		SeverityCounts: SeverityCounts{
			Low: 1,
		},
		ScanID: 999,
		// No ScanURL provided - should be injected
	}

	err := n.SendNotification(context.Background(), config, payload)
	assert.NoError(t, err)

	// Verify ScanURL was injected
	body := <-receivedPayload
	assert.Contains(t, body, "http://frontend.local/scans/999")
}

func TestSendNotification_WebhookError(t *testing.T) {
	// Create mock webhook server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	config := WebhookConfig{
		URL:         server.URL,
		MinSeverity: "Low",
	}

	payload := NotificationPayload{
		Image:      "test:latest",
		TotalVulns: 1,
		SeverityCounts: SeverityCounts{
			Low: 1,
		},
		ScanID: 1,
	}

	err := n.SendNotification(context.Background(), config, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-2xx status")
}

func TestSendNotification_InvalidURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	config := WebhookConfig{
		URL:         "://invalid-url",
		MinSeverity: "Low",
	}

	payload := NotificationPayload{
		Image:      "test:latest",
		TotalVulns: 1,
		SeverityCounts: SeverityCounts{
			Low: 1,
		},
		ScanID: 1,
	}

	err := n.SendNotification(context.Background(), config, payload)
	assert.Error(t, err)
}

func TestSendWebhook_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	n := New(logger, "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := n.sendWebhook(ctx, server.URL, map[string]string{"test": "data"})
	assert.Error(t, err)
}

func TestNew(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://test.local")

	assert.NotNil(t, n)
	assert.NotNil(t, n.logger)
	assert.NotNil(t, n.httpClient)
	assert.Equal(t, "http://test.local", n.frontendURL)
}
