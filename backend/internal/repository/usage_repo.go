package repository

import (
	"context"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/google/uuid"
)

type UsageRepository interface {
	// Gets or creates the current period's usage record for a user
	GetOrCreateCurrent(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (*models.UsageTracking, error)

	// Increments the items_processed counter for the current period
	IncrementUsage(ctx context.Context, userID uuid.UUID) error

	// Gets current period's usage for a user
	GetCurrentUsage(ctx context.Context, userID uuid.UUID) (*models.UsageTracking, error)

	// Updates the items_limit for a user's current period
	UpdateLimit(ctx context.Context, userID uuid.UUID, limit *int) error

	// Creates new periods for users whose current period has ended
	// Should be run as a cron job
	ResetExpiredPeriods(ctx context.Context) (int, error)

	// Gets usage history for a user
	GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error)
}

func GetCurrentPeriod() (start, end time.Time) {
	now := time.Now()
	start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 1, 0).Add(-time.Second) //Last second of the month
	return start, end
}
