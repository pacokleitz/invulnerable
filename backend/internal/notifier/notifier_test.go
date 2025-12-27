package notifier

import (
	"testing"

	"go.uber.org/zap"
)

func TestShouldNotify(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	tests := []struct {
		name        string
		minSeverity string
		counts      SeverityCounts
		expected    bool
	}{
		{
			name:        "Critical threshold with critical vulns",
			minSeverity: "Critical",
			counts:      SeverityCounts{Critical: 1},
			expected:    true,
		},
		{
			name:        "Critical threshold with only high vulns",
			minSeverity: "Critical",
			counts:      SeverityCounts{High: 5},
			expected:    false,
		},
		{
			name:        "High threshold with high vulns",
			minSeverity: "High",
			counts:      SeverityCounts{High: 2},
			expected:    true,
		},
		{
			name:        "High threshold with critical vulns",
			minSeverity: "High",
			counts:      SeverityCounts{Critical: 1},
			expected:    true,
		},
		{
			name:        "Medium threshold with low vulns",
			minSeverity: "Medium",
			counts:      SeverityCounts{Low: 10},
			expected:    false,
		},
		{
			name:        "Medium threshold with medium vulns",
			minSeverity: "Medium",
			counts:      SeverityCounts{Medium: 3},
			expected:    true,
		},
		{
			name:        "Low threshold with any vulns",
			minSeverity: "Low",
			counts:      SeverityCounts{Low: 1, Medium: 2},
			expected:    true,
		},
		{
			name:        "Negligible threshold with negligible vulns",
			minSeverity: "Negligible",
			counts:      SeverityCounts{Negligible: 5},
			expected:    true,
		},
		{
			name:        "No vulnerabilities",
			minSeverity: "Low",
			counts:      SeverityCounts{},
			expected:    false,
		},
		{
			name:        "Unknown severity defaults to High",
			minSeverity: "Unknown",
			counts:      SeverityCounts{High: 1},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.shouldNotify(tt.minSeverity, tt.counts)
			if result != tt.expected {
				t.Errorf("shouldNotify(%s, %+v) = %v, expected %v",
					tt.minSeverity, tt.counts, result, tt.expected)
			}
		})
	}
}

func TestGetSeverityColor(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	tests := []struct {
		name     string
		counts   SeverityCounts
		expected string
	}{
		{
			name:     "Critical vulnerabilities",
			counts:   SeverityCounts{Critical: 1},
			expected: "danger",
		},
		{
			name:     "High vulnerabilities",
			counts:   SeverityCounts{High: 2},
			expected: "warning",
		},
		{
			name:     "Medium vulnerabilities",
			counts:   SeverityCounts{Medium: 5},
			expected: "#ffcc00",
		},
		{
			name:     "Only low vulnerabilities",
			counts:   SeverityCounts{Low: 10},
			expected: "good",
		},
		{
			name:     "No vulnerabilities",
			counts:   SeverityCounts{},
			expected: "good",
		},
		{
			name:     "Mixed with critical",
			counts:   SeverityCounts{Critical: 1, High: 5, Medium: 10},
			expected: "danger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.getSeverityColor(tt.counts)
			if result != tt.expected {
				t.Errorf("getSeverityColor(%+v) = %v, expected %v",
					tt.counts, result, tt.expected)
			}
		})
	}
}

func TestGetTeamsColor(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	n := New(logger, "http://localhost:3000")

	tests := []struct {
		name     string
		counts   SeverityCounts
		expected string
	}{
		{
			name:     "Critical vulnerabilities",
			counts:   SeverityCounts{Critical: 1},
			expected: "FF0000",
		},
		{
			name:     "High vulnerabilities",
			counts:   SeverityCounts{High: 2},
			expected: "FFA500",
		},
		{
			name:     "Medium vulnerabilities",
			counts:   SeverityCounts{Medium: 5},
			expected: "FFCC00",
		},
		{
			name:     "Only low vulnerabilities",
			counts:   SeverityCounts{Low: 10},
			expected: "00FF00",
		},
		{
			name:     "No vulnerabilities",
			counts:   SeverityCounts{},
			expected: "00FF00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.getTeamsColor(tt.counts)
			if result != tt.expected {
				t.Errorf("getTeamsColor(%+v) = %v, expected %v",
					tt.counts, result, tt.expected)
			}
		})
	}
}
