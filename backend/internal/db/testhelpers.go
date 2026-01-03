package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDatabase creates a PostgreSQL testcontainer and runs migrations
func SetupTestDatabase(t *testing.T) *Database {
	t.Helper()

	ctx := context.Background()

	// Create a new container for each test
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Cleanup on test end
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Run migrations
	if err := runMigrations(connStr); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create connection
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return &Database{DB: db}
}

func runMigrations(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get path to migrations directory
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	// Discover all .up.sql migration files
	// This automatically finds all migrations, so new migrations don't require code changes
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []string
	for _, entry := range entries {
		// Match files ending with .up.sql (e.g., 001_initial_schema.up.sql)
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" &&
			len(entry.Name()) > 7 && entry.Name()[len(entry.Name())-7:] == ".up.sql" {
			migrations = append(migrations, entry.Name())
		}
	}

	// Migrations are already sorted by os.ReadDir (alphabetically)
	// Since filenames start with numbers (001_, 002_, etc.), they run in order

	// Run each migration file
	for _, migration := range migrations {
		migrationPath := filepath.Join(migrationsPath, migration)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migration, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration, err)
		}
	}

	return nil
}
