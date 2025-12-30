package models

import "time"

type Image struct {
	ID         int       `db:"id" json:"id"`
	Registry   string    `db:"registry" json:"registry"`
	Repository string    `db:"repository" json:"repository"`
	Tag        string    `db:"tag" json:"tag"`
	Digest     *string   `db:"digest" json:"digest,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type ImageWithStats struct {
	Image
	ScanCount     int        `db:"scan_count" json:"scan_count"`
	LastScanDate  *time.Time `db:"last_scan_date" json:"last_scan_date,omitempty"`
	CriticalCount int        `db:"critical_count" json:"critical_count"`
	HighCount     int        `db:"high_count" json:"high_count"`
	MediumCount   int        `db:"medium_count" json:"medium_count"`
	LowCount      int        `db:"low_count" json:"low_count"`
}

func (i *Image) FullName() string {
	if i.Registry != "" {
		return i.Registry + "/" + i.Repository + ":" + i.Tag
	}
	return i.Repository + ":" + i.Tag
}
