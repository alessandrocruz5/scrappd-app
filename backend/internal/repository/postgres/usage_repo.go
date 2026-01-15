package postgres

import (
	"context"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type usageRepository struct {
	db *pgxpool.Pool
}

// NewUsageRepository creates a new PostgreSQL usage repository
func NewUsageRepository(db *pgxpool.Pool) repository.UsageRepository {
	return &usageRepository{db: db}
}

func (r *usageRepository) GetOrCreateCurrent(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (*models.UsageTracking, error) {
	// Try to get current usage first
	usage, err := r.GetCurrentUsage(ctx, userID)
	if err == nil {
		return usage, nil
	}

	// If not found, create new record
	if appErr, ok := err.(*utils.AppError); ok && appErr.Code == utils.ErrCodeNotFound {
		periodStart, periodEnd := repository.GetCurrentPeriod()
		limit := tier.GetItemsLimit()

		query := `
			INSERT INTO usage_tracking (
				id, user_id, period_start, period_end, items_processed, items_limit
			) VALUES (
				$1, $2, $3, $4, $5, $6
			)
			RETURNING id, user_id, period_start, period_end, items_processed, items_limit, created_at, updated_at
		`

		newUsage := &models.UsageTracking{}
		err := r.db.QueryRow(
			ctx,
			query,
			uuid.New(),
			userID,
			periodStart,
			periodEnd,
			0,
			limit,
		).Scan(
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
			return nil, utils.ErrDatabase("Failed to create usage tracking", err)
		}

		return newUsage, nil
	}

	return nil, err
}

func (r *usageRepository) IncrementUsage(ctx context.Context, userID uuid.UUID) error {
	periodStart, periodEnd := repository.GetCurrentPeriod()

	query := `
		UPDATE usage_tracking
		SET items_processed = items_processed + 1
		WHERE user_id = $1 
		  AND period_start = $2 
		  AND period_end = $3
	`

	result, err := r.db.Exec(ctx, query, userID, periodStart, periodEnd)
	if err != nil {
		return utils.ErrDatabase("Failed to increment usage", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrNotFound("Usage tracking record")
	}

	return nil
}

func (r *usageRepository) GetCurrentUsage(ctx context.Context, userID uuid.UUID) (*models.UsageTracking, error) {
	periodStart, periodEnd := repository.GetCurrentPeriod()

	query := `
		SELECT 
			id, user_id, period_start, period_end, 
			items_processed, items_limit, 
			created_at, updated_at
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
		return nil, utils.ErrNotFound("Usage tracking record")
	}
	if err != nil {
		return nil, utils.ErrDatabase("Failed to get current usage", err)
	}

	return &usage, nil
}

func (r *usageRepository) UpdateLimit(ctx context.Context, userID uuid.UUID, limit *int) error {
	periodStart, periodEnd := repository.GetCurrentPeriod()

	query := `
		UPDATE usage_tracking
		SET items_limit = $1
		WHERE user_id = $2 
		  AND period_start = $3 
		  AND period_end = $4
	`

	result, err := r.db.Exec(ctx, query, limit, userID, periodStart, periodEnd)
	if err != nil {
		return utils.ErrDatabase("Failed to update usage limit", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrNotFound("Usage tracking record")
	}

	return nil
}

func (r *usageRepository) ResetExpiredPeriods(ctx context.Context) (int, error) {
	// This would be called by a cron job to create new periods for users
	// whose current period has expired

	// For now, we'll just return 0 as this will be implemented in Sprint 2
	// when we add the actual cron job
	return 0, nil
}

func (r *usageRepository) GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error) {
	query := `
		SELECT 
			id, user_id, period_start, period_end, 
			items_processed, items_limit, 
			created_at, updated_at
		FROM usage_tracking
		WHERE user_id = $1
		ORDER BY period_start DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, utils.ErrDatabase("Failed to get usage history", err)
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
			return nil, utils.ErrDatabase("Failed to scan usage record", err)
		}
		history = append(history, usage)
	}

	return history, nil
}
