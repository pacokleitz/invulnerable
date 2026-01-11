package metrics

import (
	"context"

	"github.com/invulnerable/backend/internal/db"
	"go.uber.org/zap"
)

type Service struct {
	db     *db.Database
	logger *zap.Logger
}

func New(database *db.Database, logger *zap.Logger) *Service {
	return &Service{
		db:     database,
		logger: logger,
	}
}

type DashboardMetrics struct {
	TotalImages           int            `json:"total_images"`
	TotalScans            int            `json:"total_scans"`
	TotalVulnerabilities  int            `json:"total_vulnerabilities"`
	ActiveVulnerabilities int            `json:"active_vulnerabilities"`
	SeverityCounts        SeverityCounts `json:"severity_counts"`
	RecentScans           int            `json:"recent_scans_24h"`
}

type SeverityCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

func (s *Service) GetDashboardMetrics(ctx context.Context, hasFix *bool, imageName *string) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{}

	// Prepare image name pattern for LIKE queries
	imageNamePattern := ""
	hasImageFilter := imageName != nil && *imageName != ""
	if hasImageFilter {
		imageNamePattern = "%" + *imageName + "%"
	}

	// Total images (with optional filter)
	if hasImageFilter {
		err := s.db.GetContext(ctx, &metrics.TotalImages,
			"SELECT COUNT(*) FROM images i WHERE (COALESCE(i.registry, '') || '/' || COALESCE(i.repository, '') || ':' || COALESCE(i.tag, '')) LIKE $1",
			imageNamePattern)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.db.GetContext(ctx, &metrics.TotalImages,
			"SELECT COUNT(*) FROM images i")
		if err != nil {
			return nil, err
		}
	}

	// Total scans (with optional image filter)
	if hasImageFilter {
		err := s.db.GetContext(ctx, &metrics.TotalScans,
			"SELECT COUNT(*) FROM scans s JOIN images i ON s.image_id = i.id WHERE (COALESCE(i.registry, '') || '/' || COALESCE(i.repository, '') || ':' || COALESCE(i.tag, '')) LIKE $1",
			imageNamePattern)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.db.GetContext(ctx, &metrics.TotalScans,
			"SELECT COUNT(*) FROM scans s")
		if err != nil {
			return nil, err
		}
	}

	// Build vulnerability queries with proper parameterization
	var vulnQuery string
	var vulnArgs []interface{}

	if hasImageFilter {
		// Build WHERE clause conditions
		conditions := []string{"v.status = 'active'"}
		vulnArgs = append(vulnArgs, imageNamePattern)

		if hasFix != nil {
			if *hasFix {
				conditions = append(conditions, "v.fix_version IS NOT NULL")
			} else {
				conditions = append(conditions, "v.fix_version IS NULL")
			}
		}
		conditions = append(conditions, "(COALESCE(i.registry, '') || '/' || COALESCE(i.repository, '') || ':' || COALESCE(i.tag, '')) LIKE $1")

		whereClause := ""
		for i, cond := range conditions {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += cond
		}

		// Total vulnerabilities (using scan_vulnerabilities join table)
		err := s.db.GetContext(ctx, &metrics.TotalVulnerabilities,
			"SELECT COUNT(DISTINCT v.id) FROM vulnerabilities v JOIN scan_vulnerabilities sv ON v.id = sv.vulnerability_id JOIN scans s ON sv.scan_id = s.id JOIN images i ON s.image_id = i.id WHERE "+whereClause,
			vulnArgs...)
		if err != nil {
			return nil, err
		}

		// Active vulnerabilities (same as total since we filter by status='active')
		err = s.db.GetContext(ctx, &metrics.ActiveVulnerabilities,
			"SELECT COUNT(DISTINCT v.id) FROM vulnerabilities v JOIN scan_vulnerabilities sv ON v.id = sv.vulnerability_id JOIN scans s ON sv.scan_id = s.id JOIN images i ON s.image_id = i.id WHERE "+whereClause,
			vulnArgs...)
		if err != nil {
			return nil, err
		}

		// Severity counts
		vulnQuery = `
			SELECT
				COUNT(DISTINCT CASE WHEN v.severity = 'Critical' THEN v.id END) as critical,
				COUNT(DISTINCT CASE WHEN v.severity = 'High' THEN v.id END) as high,
				COUNT(DISTINCT CASE WHEN v.severity = 'Medium' THEN v.id END) as medium,
				COUNT(DISTINCT CASE WHEN v.severity = 'Low' THEN v.id END) as low
			FROM vulnerabilities v
			JOIN scan_vulnerabilities sv ON v.id = sv.vulnerability_id
			JOIN scans s ON sv.scan_id = s.id
			JOIN images i ON s.image_id = i.id
			WHERE ` + whereClause
		err = s.db.GetContext(ctx, &metrics.SeverityCounts, vulnQuery, vulnArgs...)
		if err != nil {
			return nil, err
		}
	} else {
		// No image filter - simpler queries
		conditions := []string{"v.status = 'active'"}

		if hasFix != nil {
			if *hasFix {
				conditions = append(conditions, "v.fix_version IS NOT NULL")
			} else {
				conditions = append(conditions, "v.fix_version IS NULL")
			}
		}

		whereClause := ""
		for i, cond := range conditions {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += cond
		}

		// Total vulnerabilities
		err := s.db.GetContext(ctx, &metrics.TotalVulnerabilities,
			"SELECT COUNT(*) FROM vulnerabilities v WHERE "+whereClause)
		if err != nil {
			return nil, err
		}

		// Active vulnerabilities
		err = s.db.GetContext(ctx, &metrics.ActiveVulnerabilities,
			"SELECT COUNT(*) FROM vulnerabilities v WHERE "+whereClause)
		if err != nil {
			return nil, err
		}

		// Severity counts
		vulnQuery = `
			SELECT
				COUNT(CASE WHEN v.severity = 'Critical' THEN 1 END) as critical,
				COUNT(CASE WHEN v.severity = 'High' THEN 1 END) as high,
				COUNT(CASE WHEN v.severity = 'Medium' THEN 1 END) as medium,
				COUNT(CASE WHEN v.severity = 'Low' THEN 1 END) as low
			FROM vulnerabilities v
			WHERE ` + whereClause
		err = s.db.GetContext(ctx, &metrics.SeverityCounts, vulnQuery)
		if err != nil {
			return nil, err
		}
	}

	// Recent scans (last 24 hours)
	if hasImageFilter {
		err := s.db.GetContext(ctx, &metrics.RecentScans,
			"SELECT COUNT(*) FROM scans s JOIN images i ON s.image_id = i.id WHERE s.scan_date > NOW() - INTERVAL '24 hours' AND (COALESCE(i.registry, '') || '/' || COALESCE(i.repository, '') || ':' || COALESCE(i.tag, '')) LIKE $1",
			imageNamePattern)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.db.GetContext(ctx, &metrics.RecentScans,
			"SELECT COUNT(*) FROM scans s WHERE scan_date > NOW() - INTERVAL '24 hours'")
		if err != nil {
			return nil, err
		}
	}

	return metrics, nil
}
