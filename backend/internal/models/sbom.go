package models

import (
	"time"
)

// SBOM represents SBOM metadata stored in database
// The actual SBOM document is stored in S3 at path: scans/{scan_id}/sbom.json
type SBOM struct {
	ID        int       `db:"id" json:"id"`
	ScanID    int       `db:"scan_id" json:"scan_id"`
	Format    string    `db:"format" json:"format"` // cyclonedx or spdx
	Version   *string   `db:"version" json:"version,omitempty"`
	SizeBytes *int64    `db:"size_bytes" json:"size_bytes,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
