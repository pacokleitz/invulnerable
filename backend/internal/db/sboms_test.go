package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/invulnerable/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSBOMRepository_Create(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	version := "1.4"
	document := json.RawMessage(`{"bomFormat":"CycloneDX","specVersion":"1.4"}`)

	sbom := &models.SBOM{
		ScanID:   1,
		Format:   "cyclonedx",
		Version:  &version,
		Document: document,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at"}).
		AddRow(1, now)

	mock.ExpectQuery(`INSERT INTO sboms`).
		WithArgs(sbom.ScanID, sbom.Format, sbom.Version, sbom.Document).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), sbom)
	assert.NoError(t, err)
	assert.Equal(t, 1, sbom.ID)
	assert.NotZero(t, sbom.CreatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSBOMRepository_Create_Error(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	version := "1.4"
	document := json.RawMessage(`{"bomFormat":"CycloneDX"}`)

	sbom := &models.SBOM{
		ScanID:   1,
		Format:   "cyclonedx",
		Version:  &version,
		Document: document,
	}

	mock.ExpectQuery(`INSERT INTO sboms`).
		WithArgs(sbom.ScanID, sbom.Format, sbom.Version, sbom.Document).
		WillReturnError(sql.ErrConnDone)

	err = repo.Create(context.Background(), sbom)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSBOMRepository_GetByScanID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	now := time.Now()
	version := "1.4"
	document := json.RawMessage(`{"bomFormat":"CycloneDX","specVersion":"1.4"}`)

	rows := sqlmock.NewRows([]string{"id", "scan_id", "format", "version", "document", "created_at"}).
		AddRow(1, 1, "cyclonedx", &version, document, now)

	mock.ExpectQuery(`SELECT \* FROM sboms WHERE scan_id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	sbom, err := repo.GetByScanID(context.Background(), 1)
	assert.NoError(t, err)
	require.NotNil(t, sbom)
	assert.Equal(t, 1, sbom.ID)
	assert.Equal(t, 1, sbom.ScanID)
	assert.Equal(t, "cyclonedx", sbom.Format)
	assert.Equal(t, &version, sbom.Version)
	assert.Equal(t, document, sbom.Document)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSBOMRepository_GetByScanID_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	mock.ExpectQuery(`SELECT \* FROM sboms WHERE scan_id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	sbom, err := repo.GetByScanID(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, sbom)
	assert.Contains(t, err.Error(), "SBOM not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSBOMRepository_GetDocumentByScanID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	document := json.RawMessage(`{"bomFormat":"CycloneDX","specVersion":"1.4","components":[]}`)

	rows := sqlmock.NewRows([]string{"document"}).
		AddRow(document)

	mock.ExpectQuery(`SELECT document FROM sboms WHERE scan_id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	doc, err := repo.GetDocumentByScanID(context.Background(), 1)
	assert.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, document, doc)

	// Verify the JSON is valid
	var result map[string]interface{}
	err = json.Unmarshal(doc, &result)
	assert.NoError(t, err)
	assert.Equal(t, "CycloneDX", result["bomFormat"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSBOMRepository_GetDocumentByScanID_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	mock.ExpectQuery(`SELECT document FROM sboms WHERE scan_id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	doc, err := repo.GetDocumentByScanID(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, doc)
	assert.Contains(t, err.Error(), "SBOM not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSBOMRepository_Create_UpsertConflict(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := &Database{DB: sqlx.NewDb(mockDB, "sqlmock")}
	repo := NewSBOMRepository(db)

	version := "1.5"
	document := json.RawMessage(`{"bomFormat":"CycloneDX","specVersion":"1.5"}`)

	sbom := &models.SBOM{
		ScanID:   1,
		Format:   "cyclonedx",
		Version:  &version,
		Document: document,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at"}).
		AddRow(1, now)

	mock.ExpectQuery(`INSERT INTO sboms`).
		WithArgs(sbom.ScanID, sbom.Format, sbom.Version, sbom.Document).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), sbom)
	assert.NoError(t, err)
	assert.Equal(t, 1, sbom.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
