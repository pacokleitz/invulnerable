package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/invulnerable/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanRepository_Create(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"
	scanDate, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")

	scan := &models.Scan{
		ImageID:      1,
		ScanDate:     scanDate,
		SyftVersion:  &syftVersion,
		GrypeVersion: &grypeVersion,
		Status:       "completed",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(1, now, now)

	mock.ExpectQuery(`INSERT INTO scans`).
		WithArgs(scan.ImageID, scan.ScanDate, scan.SyftVersion, scan.GrypeVersion, scan.Status).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), scan)
	assert.NoError(t, err)
	assert.Equal(t, 1, scan.ID)
	assert.NotZero(t, scan.CreatedAt)
	assert.NotZero(t, scan.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_Create_Error(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"
	scanDate, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")

	scan := &models.Scan{
		ImageID:      1,
		ScanDate:     scanDate,
		SyftVersion:  &syftVersion,
		GrypeVersion: &grypeVersion,
		Status:       "completed",
	}

	mock.ExpectQuery(`INSERT INTO scans`).
		WithArgs(scan.ImageID, scan.ScanDate, scan.SyftVersion, scan.GrypeVersion, scan.Status).
		WillReturnError(sql.ErrConnDone)

	err = repo.Create(context.Background(), scan)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	now := time.Now()
	scanDate, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"

	rows := sqlmock.NewRows([]string{"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at"}).
		AddRow(1, 1, scanDate, &syftVersion, &grypeVersion, "completed", now, now)

	mock.ExpectQuery(`SELECT \* FROM scans WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	scan, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	require.NotNil(t, scan)
	assert.Equal(t, 1, scan.ID)
	assert.Equal(t, 1, scan.ImageID)
	assert.Equal(t, scanDate, scan.ScanDate)
	assert.Equal(t, &syftVersion, scan.SyftVersion)
	assert.Equal(t, &grypeVersion, scan.GrypeVersion)
	assert.Equal(t, "completed", scan.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetByID_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	mock.ExpectQuery(`SELECT \* FROM scans WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	scan, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, scan)
	assert.Contains(t, err.Error(), "scan not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetWithDetails(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	now := time.Now()
	scanDate, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"

	rows := sqlmock.NewRows([]string{
		"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at",
		"image_name", "vulnerability_count", "critical_count", "high_count", "medium_count", "low_count",
	}).AddRow(1, 1, scanDate, &syftVersion, &grypeVersion, "completed", now, now,
		"docker.io/library/nginx:latest", 10, 2, 3, 4, 1)

	mock.ExpectQuery(`SELECT\s+s\.\*,`).
		WithArgs(1).
		WillReturnRows(rows)

	scan, err := repo.GetWithDetails(context.Background(), 1)
	assert.NoError(t, err)
	require.NotNil(t, scan)
	assert.Equal(t, 1, scan.ID)
	assert.Equal(t, "docker.io/library/nginx:latest", scan.ImageName)
	assert.Equal(t, 10, scan.VulnerabilityCount)
	assert.Equal(t, 2, scan.CriticalCount)
	assert.Equal(t, 3, scan.HighCount)
	assert.Equal(t, 4, scan.MediumCount)
	assert.Equal(t, 1, scan.LowCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetWithDetails_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	mock.ExpectQuery(`SELECT\s+s\.\*,`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	scan, err := repo.GetWithDetails(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, scan)
	assert.Contains(t, err.Error(), "scan not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_List(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	now := time.Now()
	scanDate1, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	scanDate2, _ := time.Parse(time.RFC3339, "2024-01-14T09:00:00Z")
	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"

	rows := sqlmock.NewRows([]string{
		"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at",
		"image_name", "vulnerability_count", "critical_count", "high_count", "medium_count", "low_count",
	}).
		AddRow(1, 1, scanDate1, &syftVersion, &grypeVersion, "completed", now, now,
			"docker.io/library/nginx:latest", 10, 2, 3, 4, 1).
		AddRow(2, 2, scanDate2, &syftVersion, &grypeVersion, "completed", now, now,
			"docker.io/library/postgres:15", 5, 1, 2, 1, 1)

	mock.ExpectQuery(`SELECT\s+s\.\*,`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	scans, err := repo.List(context.Background(), 10, 0, nil)
	assert.NoError(t, err)
	require.Len(t, scans, 2)
	assert.Equal(t, 1, scans[0].ID)
	assert.Equal(t, "docker.io/library/nginx:latest", scans[0].ImageName)
	assert.Equal(t, 2, scans[1].ID)
	assert.Equal(t, "docker.io/library/postgres:15", scans[1].ImageName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_List_WithImageFilter(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	now := time.Now()
	scanDate, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	syftVersion := "0.100.0"
	grypeVersion := "0.74.0"
	imageID := 1

	rows := sqlmock.NewRows([]string{
		"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at",
		"image_name", "vulnerability_count", "critical_count", "high_count", "medium_count", "low_count",
	}).
		AddRow(1, 1, scanDate, &syftVersion, &grypeVersion, "completed", now, now,
			"docker.io/library/nginx:latest", 10, 2, 3, 4, 1)

	mock.ExpectQuery(`SELECT\s+s\.\*,.*WHERE s\.image_id`).
		WithArgs(imageID, 10, 0).
		WillReturnRows(rows)

	scans, err := repo.List(context.Background(), 10, 0, &imageID)
	assert.NoError(t, err)
	require.Len(t, scans, 1)
	assert.Equal(t, 1, scans[0].ImageID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_List_EmptyResult(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at",
		"image_name", "vulnerability_count", "critical_count", "high_count", "medium_count", "low_count",
	})

	mock.ExpectQuery(`SELECT\s+s\.\*,`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	scans, err := repo.List(context.Background(), 10, 0, nil)
	assert.NoError(t, err)
	assert.Len(t, scans, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetPreviousScan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	now := time.Now()
	scanDate, _ := time.Parse(time.RFC3339, "2024-01-14T10:00:00Z")
	syftVersion := "0.99.0"
	grypeVersion := "0.73.0"

	rows := sqlmock.NewRows([]string{"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at"}).
		AddRow(1, 1, scanDate, &syftVersion, &grypeVersion, "completed", now, now)

	mock.ExpectQuery(`SELECT \* FROM scans\s+WHERE image_id = \$1 AND scan_date < \$2`).
		WithArgs(1, "2024-01-15T10:00:00Z").
		WillReturnRows(rows)

	scan, err := repo.GetPreviousScan(context.Background(), 1, "2024-01-15T10:00:00Z")
	assert.NoError(t, err)
	require.NotNil(t, scan)
	assert.Equal(t, 1, scan.ID)
	assert.Equal(t, scanDate, scan.ScanDate)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetPreviousScan_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	mock.ExpectQuery(`SELECT \* FROM scans\s+WHERE image_id = \$1 AND scan_date < \$2`).
		WithArgs(1, "2024-01-15T10:00:00Z").
		WillReturnError(sql.ErrNoRows)

	scan, err := repo.GetPreviousScan(context.Background(), 1, "2024-01-15T10:00:00Z")
	assert.NoError(t, err) // No previous scan is not an error
	assert.Nil(t, scan)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetVulnerabilities(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	now := time.Now()
	fixVersion1 := "1.1.2"
	fixVersion2 := "7.70.0"
	url := "https://nvd.nist.gov"
	desc1 := "Critical vulnerability"
	desc2 := "High severity issue"
	notes := ""

	rows := sqlmock.NewRows([]string{
		"id", "cve_id", "package_name", "package_version", "package_type",
		"severity", "fix_version", "url", "description", "status",
		"first_detected_at", "last_seen_at", "remediation_date", "notes", "created_at", "updated_at",
	}).
		AddRow(1, "CVE-2024-0001", "openssl", "1.1.1", nil, "Critical",
			&fixVersion1, &url, &desc1, "active", now, now, nil, &notes, now, now).
		AddRow(2, "CVE-2024-0002", "curl", "7.68.0", nil, "High",
			&fixVersion2, &url, &desc2, "active", now, now, nil, &notes, now, now)

	mock.ExpectQuery(`SELECT v\.\* FROM vulnerabilities v`).
		WithArgs(1).
		WillReturnRows(rows)

	vulns, err := repo.GetVulnerabilities(context.Background(), 1)
	assert.NoError(t, err)
	require.Len(t, vulns, 2)
	assert.Equal(t, "CVE-2024-0001", vulns[0].CVEID)
	assert.Equal(t, "Critical", vulns[0].Severity)
	assert.Equal(t, "CVE-2024-0002", vulns[1].CVEID)
	assert.Equal(t, "High", vulns[1].Severity)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRepository_GetVulnerabilities_EmptyResult(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewScanRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "cve_id", "package_name", "package_version", "package_type",
		"severity", "fix_version", "url", "description", "status",
		"first_detected_at", "last_seen_at", "remediation_date", "notes", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT v\.\* FROM vulnerabilities v`).
		WithArgs(1).
		WillReturnRows(rows)

	vulns, err := repo.GetVulnerabilities(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, vulns, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}
