package repository

import (
	"context"
	"fmt"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type usageRepository struct {
	db *pgxpool.Pool
}

// NewUsageRepository creates a new PostgreSQL usage repository
func NewUsageRepository(db *pgxpool.Pool) UsageRepository {
	return &usageRepository{db: db}
}

// GetOrCreateCurrent gets or creates the current period's usage record
func (r *usageRepository) GetOrCreateCurrent(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (*models.UsageTracking, error) {
	// First, try to get current usage
	usage, err := r.GetCurrentUsage(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	// If exists, return it
	if usage != nil {
		return usage, nil
	}

	// Otherwise, create new record for current period
	periodStart, periodEnd := GetCurrentPeriod()
	limit := tier.GetItemsLimit()

	query := `
		INSERT INTO usage_tracking (user_id, period_start, period_end, items_processed, items_limit)
		VALUES ($1, $2, $3, 0, $4)
		RETURNING id, user_id, period_start, period_end, items_processed, items_limit, created_at, updated_at
	`

	var newUsage models.UsageTracking
	err = r.db.QueryRow(ctx, query, userID, periodStart, periodEnd, limit).Scan(
		&newUsage.ID,
		&newUsage.UserID,
		&newUsage.PeriodStart,
		&newUsage.PeriodEnd,
		&newUsage.ItemsProcessed,
		&newUsage.ItemsLimit,
		&newUsage.CreatedAt,
		&newUsage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	return &newUsage, nil
}

// IncrementUsage increments the items_processed counter
func (r *usageRepository) IncrementUsage(ctx context.Context, userID uuid.UUID) error {
	periodStart, periodEnd := GetCurrentPeriod()

	query := `
		UPDATE usage_tracking
		SET items_processed = items_processed + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 
		AND period_start = $2
		AND period_end = $3
	`

	result, err := r.db.Exec(ctx, query, userID, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no usage record found for current period")
	}

	return nil
}

// GetCurrentUsage gets the current period's usage record
func (r *usageRepository) GetCurrentUsage(ctx context.Context, userID uuid.UUID) (*models.UsageTracking, error) {
	periodStart, periodEnd := GetCurrentPeriod()

	query := `
		SELECT id, user_id, period_start, period_end, items_processed, items_limit, created_at, updated_at
		FROM usage_tracking
		WHERE user_id = $1 
		AND period_start = $2
		AND period_end = $3
	`

	var usage models.UsageTracking
	err := r.db.QueryRow(ctx, query, userID, periodStart, periodEnd).Scan(
		&usage.ID,
		&usage.UserID,
		&usage.PeriodStart,
		&usage.PeriodEnd,
		&usage.ItemsProcessed,
		&usage.ItemsLimit,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // No record found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	return &usage, nil
}

// UpdateLimit updates the items_limit for current period
func (r *usageRepository) UpdateLimit(ctx context.Context, userID uuid.UUID, limit *int) error {
	periodStart, periodEnd := GetCurrentPeriod()

	query := `
		UPDATE usage_tracking
		SET items_limit = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2 
		AND period_start = $3
		AND period_end = $4
	`

	result, err := r.db.Exec(ctx, query, limit, userID, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to update limit: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no usage record found for current period")
	}

	return nil
}

// ResetExpiredPeriods creates new periods for users whose current period has ended
func (r *usageRepository) ResetExpiredPeriods(ctx context.Context) (int, error) {
	periodStart, periodEnd := GetCurrentPeriod()

	query := `
		INSERT INTO usage_tracking (user_id, period_start, period_end, items_processed, items_limit)
		SELECT 
			u.id as user_id,
			$1 as period_start,
			$2 as period_end,
			0 as items_processed,
			CASE 
				WHEN u.subscription_tier = 'pro' THEN NULL
				ELSE 5
			END as items_limit
		FROM users u
		WHERE NOT EXISTS (
			SELECT 1 FROM usage_tracking ut
			WHERE ut.user_id = u.id
			AND ut.period_start = $1
			AND ut.period_end = $2
		)
	`

	result, err := r.db.Exec(ctx, query, periodStart, periodEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to reset expired periods: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetUsageHistory retrieves usage history for a user
func (r *usageRepository) GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error) {
	if limit <= 0 {
		limit = 12 // Default to last 12 months
	}

	query := `
		SELECT id, user_id, period_start, period_end, items_processed, items_limit, created_at, updated_at
		FROM usage_tracking
		WHERE user_id = $1
		ORDER BY period_start DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage history: %w", err)
	}
	defer rows.Close()

	var history []models.UsageTracking
	for rows.Next() {
		var usage models.UsageTracking
		err := rows.Scan(
			&usage.ID,
			&usage.UserID,
			&usage.PeriodStart,
			&usage.PeriodEnd,
			&usage.ItemsProcessed,
			&usage.ItemsLimit,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage record: %w", err)
		}
		history = append(history, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage records: %w", err)
	}

	return history, nil
}
