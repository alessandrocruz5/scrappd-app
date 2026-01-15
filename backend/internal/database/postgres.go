package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type DB struct {
	Pool   *pgxpool.Pool
	Logger *logrus.Logger
}

// NewDB creates a new database connection pool
func NewDB(dsn string, logger *logrus.Logger) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse the connection string
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")

	return &DB{
		Pool:   pool,
		Logger: logger,
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
	db.Logger.Info("Database connection closed")
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
