package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/invulnerable/backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImageName(t *testing.T) {
	tests := []struct {
		name             string
		fullName         string
		expectedRegistry string
		expectedRepo     string
		expectedTag      string
	}{
		{
			name:             "docker hub with tag",
			fullName:         "nginx:latest",
			expectedRegistry: "docker.io",
			expectedRepo:     "nginx",
			expectedTag:      "latest",
		},
		{
			name:             "docker hub library",
			fullName:         "library/nginx:alpine",
			expectedRegistry: "docker.io",
			expectedRepo:     "library/nginx",
			expectedTag:      "alpine",
		},
		{
			name:             "gcr with tag",
			fullName:         "gcr.io/project/image:v1.0",
			expectedRegistry: "gcr.io",
			expectedRepo:     "project/image",
			expectedTag:      "v1.0",
		},
		{
			name:             "no tag defaults to latest",
			fullName:         "nginx",
			expectedRegistry: "docker.io",
			expectedRepo:     "nginx",
			expectedTag:      "latest",
		},
		{
			name:             "localhost registry",
			fullName:         "localhost:5000/myimage:dev",
			expectedRegistry: "localhost:5000",
			expectedRepo:     "myimage",
			expectedTag:      "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, repo, tag := parseImageName(tt.fullName)
			assert.Equal(t, tt.expectedRegistry, registry)
			assert.Equal(t, tt.expectedRepo, repo)
			assert.Equal(t, tt.expectedTag, tag)
		})
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CRITICAL", "Critical"},
		{"critical", "Critical"},
		{"HIGH", "High"},
		{"high", "High"},
		{"MEDIUM", "Medium"},
		{"medium", "Medium"},
		{"LOW", "Low"},
		{"low", "Low"},
		{"unknown", "Unknown"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeSeverity(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScanHandler_CreateScan_ValidRequest(t *testing.T) {
	// This is a simplified test - in a real scenario you'd mock the repositories
	e := echo.New()

	// Create a sample Grype result
	grypeResult := models.GrypeResult{
		Matches: []models.GrypeMatch{
			{
				Vulnerability: models.GrypeVulnerability{
					ID:          "CVE-2023-1234",
					Severity:    "High",
					Description: "Test vulnerability",
				},
				Artifact: models.GrypeArtifact{
					Name:    "openssl",
					Version: "1.1.1",
					Type:    "deb",
				},
			},
		},
		Descriptor: models.GrypeDescriptor{
			Name:    "grype",
			Version: "0.65.0",
		},
	}

	sbom := json.RawMessage(`{"bomFormat": "CycloneDX"}`)

	reqBody := ScanRequest{
		Image:       "nginx:latest",
		GrypeResult: grypeResult,
		SBOM:        sbom,
		SBOMFormat:  "cyclonedx",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Note: This would need proper mocking of repositories to work fully
	// For now, this tests the request parsing
	var parsedReq ScanRequest
	err = c.Bind(&parsedReq)
	assert.NoError(t, err)
	assert.Equal(t, "nginx:latest", parsedReq.Image)
	assert.Equal(t, "cyclonedx", parsedReq.SBOMFormat)
	assert.Len(t, parsedReq.GrypeResult.Matches, 1)
}
