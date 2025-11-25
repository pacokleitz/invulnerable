package metrics

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/invulnerable/backend/internal/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*db.Database, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	database := &db.Database{DB: sqlxDB}

	return database, mock
}

func TestService_GetDashboardMetrics(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	// Mock total images count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	// Mock total scans count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

	// Mock total vulnerabilities count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

	// Mock active vulnerabilities count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(75))

	// Mock severity counts
	mock.ExpectQuery(`SELECT\s+COUNT\(CASE WHEN severity = 'Critical'`).
		WillReturnRows(sqlmock.NewRows([]string{"critical", "high", "medium", "low"}).
			AddRow(10, 20, 30, 15))

	// Mock recent scans (last 24 hours)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans WHERE scan_date`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock vulnerability trend (last 30 days)
	mock.ExpectQuery(`SELECT\s+DATE\(scan_date\) as date`).
		WillReturnRows(sqlmock.NewRows([]string{"date", "count"}).
			AddRow("2024-01-15", 25).
			AddRow("2024-01-14", 30).
			AddRow("2024-01-13", 28))

	metrics, err := service.GetDashboardMetrics(context.Background())

	require.NoError(t, err)
	require.NotNil(t, metrics)

	assert.Equal(t, 10, metrics.TotalImages)
	assert.Equal(t, 50, metrics.TotalScans)
	assert.Equal(t, 100, metrics.TotalVulnerabilities)
	assert.Equal(t, 75, metrics.ActiveVulnerabilities)
	assert.Equal(t, 10, metrics.SeverityCounts.Critical)
	assert.Equal(t, 20, metrics.SeverityCounts.High)
	assert.Equal(t, 30, metrics.SeverityCounts.Medium)
	assert.Equal(t, 15, metrics.SeverityCounts.Low)
	assert.Equal(t, 5, metrics.RecentScans)
	assert.Len(t, metrics.VulnerabilityTrend, 3)
	assert.Equal(t, "2024-01-15", metrics.VulnerabilityTrend[0].Date)
	assert.Equal(t, 25, metrics.VulnerabilityTrend[0].Count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_EmptyTrend(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	// Mock all counts
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT\s+COUNT\(CASE WHEN severity = 'Critical'`).
		WillReturnRows(sqlmock.NewRows([]string{"critical", "high", "medium", "low"}).
			AddRow(0, 0, 0, 0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans WHERE scan_date`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock empty vulnerability trend
	mock.ExpectQuery(`SELECT\s+DATE\(scan_date\) as date`).
		WillReturnRows(sqlmock.NewRows([]string{"date", "count"}))

	metrics, err := service.GetDashboardMetrics(context.Background())

	require.NoError(t, err)
	require.NotNil(t, metrics)

	assert.Equal(t, 0, metrics.TotalImages)
	assert.Equal(t, 0, metrics.TotalScans)
	assert.Equal(t, 0, metrics.TotalVulnerabilities)
	assert.Equal(t, 0, metrics.ActiveVulnerabilities)
	assert.Len(t, metrics.VulnerabilityTrend, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnImagesCount(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	// Mock error on first query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnScansCount(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	// Mock error on second query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnVulnerabilitiesCount(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

	// Mock error on third query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnActiveCount(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

	// Mock error on active vulnerabilities
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities WHERE status = 'active'`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnSeverityCounts(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(75))

	// Mock error on severity counts
	mock.ExpectQuery(`SELECT\s+COUNT\(CASE WHEN severity = 'Critical'`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnRecentScans(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(75))
	mock.ExpectQuery(`SELECT\s+COUNT\(CASE WHEN severity = 'Critical'`).
		WillReturnRows(sqlmock.NewRows([]string{"critical", "high", "medium", "low"}).
			AddRow(10, 20, 30, 15))

	// Mock error on recent scans
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans WHERE scan_date`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetDashboardMetrics_ErrorOnTrend(t *testing.T) {
	database, mock := setupMockDB(t)
	defer database.Close()

	service := New(database)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM vulnerabilities WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(75))
	mock.ExpectQuery(`SELECT\s+COUNT\(CASE WHEN severity = 'Critical'`).
		WillReturnRows(sqlmock.NewRows([]string{"critical", "high", "medium", "low"}).
			AddRow(10, 20, 30, 15))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM scans WHERE scan_date`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock error on vulnerability trend
	mock.ExpectQuery(`SELECT\s+DATE\(scan_date\) as date`).
		WillReturnError(sqlmock.ErrCancelled)

	metrics, err := service.GetDashboardMetrics(context.Background())

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}
