// internal/repository/test_helpers_test.go
package repository

import (
	"context"
	"os"
	"testing"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://scrappd_app:scrappd-go@localhost:5432/scrappd?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Create a test logger (discards output to keep tests clean)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.WarnLevel) // Only show warnings and errors in tests

	return &database.DB{
		Pool:   pool,
		Logger: logger,
	}
}

func cleanupTestDB(t *testing.T, db *database.DB) {
	t.Helper()

	ctx := context.Background()

	// Clean up test data in reverse order of dependencies
	_, _ = db.Pool.Exec(ctx, "DELETE FROM content.usage_tracking")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM content.items")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM auth.users")

	db.Pool.Close()
}
