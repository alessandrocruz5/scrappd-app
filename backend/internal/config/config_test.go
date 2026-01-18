package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test default server config
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "development", cfg.Server.Environment)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)

	// Test default database config
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "scrappd_app", cfg.Database.User)

	// Test default ML service config
	assert.Equal(t, "http://localhost:8000", cfg.MLService.BaseURL)
	assert.Equal(t, 120*time.Second, cfg.MLService.Timeout)
	assert.Equal(t, 3, cfg.MLService.MaxRetries)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set custom environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("ML_SERVICE_URL", "http://ml-service:8000")
	os.Setenv("ML_SERVICE_TIMEOUT", "60s")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")
	defer os.Clearenv()

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "production", cfg.Server.Environment)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "http://ml-service:8000", cfg.MLService.BaseURL)
	assert.Equal(t, 60*time.Second, cfg.MLService.Timeout)
}

func TestDatabaseDSN(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	defer os.Clearenv()

	cfg, err := Load()
	require.NoError(t, err)

	expectedDSN := "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expectedDSN, cfg.Database.DSN)
}
