package metrics

import (
	"context"

	"github.com/invulnerable/backend/internal/db"
)

type Service struct {
	db *db.Database
}

func New(database *db.Database) *Service {
	return &Service{db: database}
}

type DashboardMetrics struct {
	TotalImages            int                    `json:"total_images"`
	TotalScans             int                    `json:"total_scans"`
	TotalVulnerabilities   int                    `json:"total_vulnerabilities"`
	ActiveVulnerabilities  int                    `json:"active_vulnerabilities"`
	SeverityCounts         SeverityCounts         `json:"severity_counts"`
	RecentScans            int                    `json:"recent_scans_24h"`
	VulnerabilityTrend     []VulnerabilityTrendPoint `json:"vulnerability_trend"`
}

type SeverityCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

type VulnerabilityTrendPoint struct {
	Date   string `json:"date"`
	Count  int    `json:"count"`
}

func (s *Service) GetDashboardMetrics(ctx context.Context) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{}

	// Total images
	err := s.db.GetContext(ctx, &metrics.TotalImages, "SELECT COUNT(*) FROM images")
	if err != nil {
		return nil, err
	}

	// Total scans
	err = s.db.GetContext(ctx, &metrics.TotalScans, "SELECT COUNT(*) FROM scans")
	if err != nil {
		return nil, err
	}

	// Total vulnerabilities
	err = s.db.GetContext(ctx, &metrics.TotalVulnerabilities, "SELECT COUNT(*) FROM vulnerabilities")
	if err != nil {
		return nil, err
	}

	// Active vulnerabilities
	err = s.db.GetContext(ctx, &metrics.ActiveVulnerabilities,
		"SELECT COUNT(*) FROM vulnerabilities WHERE status = 'active'")
	if err != nil {
		return nil, err
	}

	// Severity counts
	query := `
		SELECT
			COUNT(CASE WHEN severity = 'Critical' THEN 1 END) as critical,
			COUNT(CASE WHEN severity = 'High' THEN 1 END) as high,
			COUNT(CASE WHEN severity = 'Medium' THEN 1 END) as medium,
			COUNT(CASE WHEN severity = 'Low' THEN 1 END) as low
		FROM vulnerabilities
		WHERE status = 'active'
	`
	err = s.db.GetContext(ctx, &metrics.SeverityCounts, query)
	if err != nil {
		return nil, err
	}

	// Recent scans (last 24 hours)
	err = s.db.GetContext(ctx, &metrics.RecentScans,
		"SELECT COUNT(*) FROM scans WHERE scan_date > NOW() - INTERVAL '24 hours'")
	if err != nil {
		return nil, err
	}

	// Vulnerability trend (last 30 days)
	metrics.VulnerabilityTrend = []VulnerabilityTrendPoint{}
	trendQuery := `
		SELECT
			DATE(scan_date) as date,
			COUNT(DISTINCT sv.vulnerability_id) as count
		FROM scans s
		LEFT JOIN scan_vulnerabilities sv ON sv.scan_id = s.id
		WHERE s.scan_date > NOW() - INTERVAL '30 days'
		GROUP BY DATE(scan_date)
		ORDER BY date DESC
	`
	err = s.db.SelectContext(ctx, &metrics.VulnerabilityTrend, trendQuery)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
