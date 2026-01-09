package database

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	dsn := "host=localhost port=5432 user=scrappd_app password=scrappd-go dbname=scrappd sslmode=disable"

	db, err := NewDB(dsn, logger)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Health(ctx)
	assert.NoError(t, err)
}

func TestNewDB_InvalidDSN(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	db, err := NewDB("invalid-dsn", logger)
	assert.Error(t, err)
	assert.Nil(t, db)
}
