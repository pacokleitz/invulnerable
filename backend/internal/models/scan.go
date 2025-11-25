package models

import "time"

type Scan struct {
	ID           int       `db:"id" json:"id"`
	ImageID      int       `db:"image_id" json:"image_id"`
	ScanDate     time.Time `db:"scan_date" json:"scan_date"`
	SyftVersion  *string   `db:"syft_version" json:"syft_version,omitempty"`
	GrypeVersion *string   `db:"grype_version" json:"grype_version,omitempty"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type ScanWithDetails struct {
	Scan
	ImageName          string `db:"image_name" json:"image_name"`
	VulnerabilityCount int    `db:"vulnerability_count" json:"vulnerability_count"`
	CriticalCount      int    `db:"critical_count" json:"critical_count"`
	HighCount          int    `db:"high_count" json:"high_count"`
	MediumCount        int    `db:"medium_count" json:"medium_count"`
	LowCount           int    `db:"low_count" json:"low_count"`
}

type ScanDiff struct {
	ScanID           int              `json:"scan_id"`
	PreviousScanID   int              `json:"previous_scan_id"`
	NewVulns         []Vulnerability  `json:"new_vulnerabilities"`
	FixedVulns       []Vulnerability  `json:"fixed_vulnerabilities"`
	PersistentVulns  []Vulnerability  `json:"persistent_vulnerabilities"`
	Summary          ScanDiffSummary  `json:"summary"`
}

type ScanDiffSummary struct {
	NewCount        int `json:"new_count"`
	FixedCount      int `json:"fixed_count"`
	PersistentCount int `json:"persistent_count"`
}
