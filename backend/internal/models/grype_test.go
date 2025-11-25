package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrypeResult_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"matches": [
			{
				"vulnerability": {
					"id": "CVE-2023-1234",
					"severity": "High",
					"description": "Test vulnerability"
				},
				"artifact": {
					"name": "openssl",
					"version": "1.1.1",
					"type": "deb"
				},
				"matchDetails": []
			}
		],
		"descriptor": {
			"name": "grype",
			"version": "0.65.0"
		}
	}`

	var result GrypeResult
	err := json.Unmarshal([]byte(jsonData), &result)

	require.NoError(t, err)
	assert.Len(t, result.Matches, 1)
	assert.Equal(t, "CVE-2023-1234", result.Matches[0].Vulnerability.ID)
	assert.Equal(t, "High", result.Matches[0].Vulnerability.Severity)
	assert.Equal(t, "openssl", result.Matches[0].Artifact.Name)
	assert.Equal(t, "grype", result.Descriptor.Name)
	assert.Equal(t, "0.65.0", result.Descriptor.Version)
}

func TestGrypeVulnerability_WithFix(t *testing.T) {
	vuln := GrypeVulnerability{
		ID:       "CVE-2023-5678",
		Severity: "Critical",
		Fix: &GrypeFix{
			Versions: []string{"1.2.3", "1.2.4"},
			State:    "fixed",
		},
	}

	assert.NotNil(t, vuln.Fix)
	assert.Len(t, vuln.Fix.Versions, 2)
	assert.Equal(t, "1.2.3", vuln.Fix.Versions[0])
}
