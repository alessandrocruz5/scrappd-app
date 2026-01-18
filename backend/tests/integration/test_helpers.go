//go:build integration
// +build integration

package integration

import (
	"context"
	"io"
	"testing"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection.
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := database.NewDB(cfg.Database.DSN, logger)
	require.NoError(t, err)

	return db
}

// cleanupTestUser removes test user from database.
func cleanupTestUser(t *testing.T, db *database.DB, email string) {
	t.Helper()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, "DELETE FROM auth.users WHERE email = $1", email)
	require.NoError(t, err)
}
