package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Get database connection details from environment or use defaults
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "scrappd_user")
	password := getEnvOrDefault("DB_PASSWORD", "scrappd_password")
	dbname := getEnvOrDefault("DB_NAME", "scrappd_db")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "Failed to connect to test database")

	// Verify connection
	err = pool.Ping(ctx)
	require.NoError(t, err, "Failed to ping test database")

	return pool
}

// cleanupTestDB cleans up test data and closes the connection
func cleanupTestDB(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()

	// Clean up test data
	tables := []string{"items", "usage_tracking"}
	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to clean up table %s: %v", table, err)
		}
	}

	// Close connection
	db.Close()
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
