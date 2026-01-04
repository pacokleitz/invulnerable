package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/invulnerable/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFilterVulnerabilitiesByOnlyFixable_Mixed(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-mixed.json")

	// Test filtering with onlyFixable=true
	filtered := filterMatchesByOnlyFixable(grypeResult.Matches, true)

	// Should only include CVEs with fixes: CVE-2024-1234 and CVE-2024-9999
	assert.Len(t, filtered, 2, "should only include vulnerabilities with fixes")

	cveIDs := make([]string, len(filtered))
	for i, match := range filtered {
		cveIDs[i] = match.Vulnerability.ID
	}

	assert.Contains(t, cveIDs, "CVE-2024-1234", "should include Critical with fix")
	assert.Contains(t, cveIDs, "CVE-2024-9999", "should include Medium with fix")
	assert.NotContains(t, cveIDs, "CVE-2024-5678", "should exclude High without fix")
	assert.NotContains(t, cveIDs, "CVE-2024-1111", "should exclude Low without fix")
}

func TestFilterVulnerabilitiesByOnlyFixable_AllFixed(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-all-fixed.json")

	// Test with onlyFixable=true
	filtered := filterMatchesByOnlyFixable(grypeResult.Matches, true)

	// All CVEs have fixes, so all should pass
	assert.Len(t, filtered, 2, "should include all fixed vulnerabilities")
	assert.Len(t, filtered, len(grypeResult.Matches), "should not filter any")
}

func TestFilterVulnerabilitiesByOnlyFixable_AllUnfixed(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-all-unfixed.json")

	// Test with onlyFixable=true
	filtered := filterMatchesByOnlyFixable(grypeResult.Matches, true)

	// No CVEs have fixes, so all should be filtered out
	assert.Len(t, filtered, 0, "should filter out all unfixed vulnerabilities")
}

func TestFilterVulnerabilitiesByOnlyFixable_Disabled(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-mixed.json")

	// Test with onlyFixable=false (disabled filtering)
	filtered := filterMatchesByOnlyFixable(grypeResult.Matches, false)

	// Should include all CVEs when filtering is disabled
	assert.Len(t, filtered, len(grypeResult.Matches), "should not filter when onlyFixable=false")
	assert.Len(t, filtered, 4, "should include all 4 CVEs")
}

func TestSeverityCounts_WithFixtures(t *testing.T) {
	grypeResult := loadGrypeFixture(t, "grype-output-mixed.json")

	// Count all vulnerabilities
	counts := countBySeverity(grypeResult.Matches)
	assert.Equal(t, 1, counts["Critical"], "should have 1 Critical")
	assert.Equal(t, 1, counts["High"], "should have 1 High")
	assert.Equal(t, 1, counts["Medium"], "should have 1 Medium")
	assert.Equal(t, 1, counts["Low"], "should have 1 Low")

	// Count only fixed vulnerabilities
	fixedMatches := filterMatchesByOnlyFixable(grypeResult.Matches, true)
	fixedCounts := countBySeverity(fixedMatches)
	assert.Equal(t, 1, fixedCounts["Critical"], "should have 1 fixed Critical")
	assert.Equal(t, 0, fixedCounts["High"], "should have 0 fixed High")
	assert.Equal(t, 1, fixedCounts["Medium"], "should have 1 fixed Medium")
	assert.Equal(t, 0, fixedCounts["Low"], "should have 0 fixed Low")
}

// Helper functions to extract from scan processing logic

func filterMatchesByOnlyFixable(matches []models.GrypeMatch, onlyFixable bool) []models.GrypeMatch {
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

func countBySeverity(matches []models.GrypeMatch) map[string]int {
	counts := make(map[string]int)
	for _, match := range matches {
		counts[match.Vulnerability.Severity]++
	}
	return counts
}
