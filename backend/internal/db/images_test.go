package db

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/invulnerable/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*Database, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db := &Database{sqlxDB}

	return db, mock
}

func TestImageRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	now := time.Now()
	img := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}

	mock.ExpectQuery(`INSERT INTO images`).
		WithArgs("docker.io", "library/nginx", "latest", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))

	err := repo.Create(context.Background(), img)

	assert.NoError(t, err)
	assert.Equal(t, 1, img.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "registry", "repository", "tag", "digest", "created_at", "updated_at"}).
		AddRow(1, "docker.io", "library/nginx", "latest", nil, now, now)

	mock.ExpectQuery(`SELECT \* FROM images WHERE id`).
		WithArgs(1).
		WillReturnRows(rows)

	img, err := repo.GetByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, 1, img.ID)
	assert.Equal(t, "docker.io", img.Registry)
	assert.Equal(t, "library/nginx", img.Repository)
	assert.Equal(t, "latest", img.Tag)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_List(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "registry", "repository", "tag", "digest", "created_at", "updated_at",
		"scan_count", "last_scan_date", "critical_count", "high_count", "medium_count", "low_count",
	}).
		AddRow(1, "docker.io", "library/nginx", "latest", nil, now, now, 5, now, 1, 2, 3, 4).
		AddRow(2, "docker.io", "library/alpine", "3.18", nil, now, now, 3, now, 0, 1, 2, 0)

	mock.ExpectQuery(`SELECT`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	images, err := repo.List(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Len(t, images, 2)
	assert.Equal(t, "library/nginx", images[0].Repository)
	assert.Equal(t, 5, images[0].ScanCount)
	assert.Equal(t, 1, images[0].CriticalCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetByName(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	now := time.Now()
	digest := "sha256:abc123"
	rows := sqlmock.NewRows([]string{"id", "registry", "repository", "tag", "digest", "created_at", "updated_at"}).
		AddRow(1, "docker.io", "library/nginx", "latest", &digest, now, now)

	mock.ExpectQuery(`SELECT \* FROM images WHERE registry = \$1 AND repository = \$2 AND tag = \$3`).
		WithArgs("docker.io", "library/nginx", "latest").
		WillReturnRows(rows)

	img, err := repo.GetByName(context.Background(), "docker.io", "library/nginx", "latest")

	require.NoError(t, err)
	require.NotNil(t, img)
	assert.Equal(t, 1, img.ID)
	assert.Equal(t, "docker.io", img.Registry)
	assert.Equal(t, "library/nginx", img.Repository)
	assert.Equal(t, "latest", img.Tag)
	assert.Equal(t, &digest, img.Digest)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetByName_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	mock.ExpectQuery(`SELECT \* FROM images WHERE registry = \$1 AND repository = \$2 AND tag = \$3`).
		WithArgs("docker.io", "library/notfound", "latest").
		WillReturnError(sqlmock.ErrCancelled)

	img, err := repo.GetByName(context.Background(), "docker.io", "library/notfound", "latest")

	assert.Error(t, err)
	assert.Nil(t, img)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetByName_NoRows(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	// Test when no rows are returned - GetByName returns nil, nil in this case
	mock.ExpectQuery(`SELECT \* FROM images WHERE registry = \$1 AND repository = \$2 AND tag = \$3`).
		WithArgs("docker.io", "library/noexist", "latest").
		WillReturnRows(sqlmock.NewRows([]string{"id", "registry", "repository", "tag", "digest", "created_at", "updated_at"}))

	img, err := repo.GetByName(context.Background(), "docker.io", "library/noexist", "latest")

	assert.NoError(t, err)
	assert.Nil(t, img)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetScanHistory(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

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
		AddRow(2, 1, scanDate2, &syftVersion, &grypeVersion, "completed", now, now,
			"docker.io/library/nginx:latest", 8, 1, 2, 3, 2)

	mock.ExpectQuery(`SELECT\s+s\.\*,`).
		WithArgs(1, 10).
		WillReturnRows(rows)

	scans, err := repo.GetScanHistory(context.Background(), 1, 10)

	require.NoError(t, err)
	assert.Len(t, scans, 2)
	assert.Equal(t, 1, scans[0].ID)
	assert.Equal(t, "docker.io/library/nginx:latest", scans[0].ImageName)
	assert.Equal(t, 10, scans[0].VulnerabilityCount)
	assert.Equal(t, 2, scans[0].CriticalCount)
	assert.Equal(t, 2, scans[1].ID)
	assert.Equal(t, 8, scans[1].VulnerabilityCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetScanHistory_EmptyResult(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewImageRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "image_id", "scan_date", "syft_version", "grype_version", "status", "created_at", "updated_at",
		"image_name", "vulnerability_count", "critical_count", "high_count", "medium_count", "low_count",
	})

	mock.ExpectQuery(`SELECT\s+s\.\*,`).
		WithArgs(999, 10).
		WillReturnRows(rows)

	scans, err := repo.GetScanHistory(context.Background(), 999, 10)

	require.NoError(t, err)
	assert.Len(t, scans, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
