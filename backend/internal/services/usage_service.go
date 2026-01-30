package services

import (
	"context"
	"fmt"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
)

// UsageService handles usage tracking business logic
type UsageService interface {
	// CheckAndIncrementUsage checks if user can process and increments if allowed
	CheckAndIncrementUsage(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (bool, *models.UsageStats, error)

	// GetCurrentUsageStats retrieves current usage stats for a user
	GetCurrentUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, error)

	// GetUsageHistory retrieves usage history for a user
	GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error)

	// UpdateSubscriptionTier updates the usage limit when subscription tier changes
	UpdateSubscriptionTier(ctx context.Context, userID uuid.UUID, newTier models.SubscriptionTier) error

	// ResetExpiredPeriods creates new periods for users (cron job)
	ResetExpiredPeriods(ctx context.Context) (int, error)

	// InitializeUsageForNewUser creates initial usage record for new user
	InitializeUsageForNewUser(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) error

	// GetRateLimitHeaders generates rate limit headers for API responses
	GetRateLimitHeaders(stats *models.UsageStats) map[string]string
}

type usageService struct {
	usageRepo repository.UsageRepository
}

// NewUsageService creates a new usage service
func NewUsageService(usageRepo repository.UsageRepository) UsageService {
	return &usageService{
		usageRepo: usageRepo,
	}
}

// CheckAndIncrementUsage checks if user can process and increments if allowed
func (s *usageService) CheckAndIncrementUsage(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (bool, *models.UsageStats, error) {
	// Get or create current usage record
	usage, err := s.usageRepo.GetOrCreateCurrent(ctx, userID, tier)
	if err != nil {
		return false, nil, utils.ErrDatabase("Failed to check usage", err)
	}

	// Build usage stats
	stats := s.buildUsageStats(usage, tier)

	// Pro users have unlimited processing
	if tier.IsUnlimited() {
		// Increment for tracking purposes
		if err := s.usageRepo.IncrementUsage(ctx, userID); err != nil {
			// Log error but don't fail the request
			// Pro users should always be able to process
		}
		return true, stats, nil
	}

	// Check if free user has remaining quota
	if usage.ItemsLimit == nil {
		return false, stats, utils.ErrInternalServer("Usage limit is nil for non-pro user", nil)
	}

	if usage.ItemsProcessed >= *usage.ItemsLimit {
		return false, stats, nil // Quota exceeded
	}

	// Increment usage
	if err := s.usageRepo.IncrementUsage(ctx, userID); err != nil {
		return false, stats, utils.ErrDatabase("Failed to increment usage", err)
	}

	// Update stats to reflect the increment
	stats.ItemsProcessed++
	if stats.ItemsRemaining != nil {
		remaining := *stats.ItemsRemaining - 1
		if remaining < 0 {
			remaining = 0
		}
		stats.ItemsRemaining = &remaining
	}

	return true, stats, nil
}

// GetCurrentUsageStats retrieves current usage stats
func (s *usageService) GetCurrentUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, error) {
	usage, err := s.usageRepo.GetCurrentUsage(ctx, userID)
	if err != nil {
		return nil, utils.ErrDatabase("Failed to get current usage", err)
	}

	if usage == nil {
		return nil, utils.ErrNotFound("Usage record")
	}

	// Determine tier from usage record
	tier := models.TierFree
	if usage.ItemsLimit == nil {
		tier = models.TierPro
	}

	stats := s.buildUsageStats(usage, tier)
	return stats, nil
}

// GetUsageHistory retrieves usage history
func (s *usageService) GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error) {
	history, err := s.usageRepo.GetUsageHistory(ctx, userID, limit)
	if err != nil {
		return nil, utils.ErrDatabase("Failed to get usage history", err)
	}

	return history, nil
}

// UpdateSubscriptionTier updates the usage limit when tier changes
func (s *usageService) UpdateSubscriptionTier(ctx context.Context, userID uuid.UUID, newTier models.SubscriptionTier) error {
	newLimit := newTier.GetItemsLimit()

	err := s.usageRepo.UpdateLimit(ctx, userID, newLimit)
	if err != nil {
		return utils.ErrDatabase("Failed to update subscription tier", err)
	}

	return nil
}

// ResetExpiredPeriods creates new periods for users
func (s *usageService) ResetExpiredPeriods(ctx context.Context) (int, error) {
	count, err := s.usageRepo.ResetExpiredPeriods(ctx)
	if err != nil {
		return 0, utils.ErrDatabase("Failed to reset expired periods", err)
	}

	return count, nil
}

// InitializeUsageForNewUser creates initial usage record
func (s *usageService) InitializeUsageForNewUser(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) error {
	// GetOrCreateCurrent will handle creation if not exists
	_, err := s.usageRepo.GetOrCreateCurrent(ctx, userID, tier)
	if err != nil {
		return utils.ErrDatabase("Failed to initialize usage", err)
	}

	return nil
}

// GetRateLimitHeaders generates rate limit headers for API responses
func (s *usageService) GetRateLimitHeaders(stats *models.UsageStats) map[string]string {
	headers := make(map[string]string)

	if stats.IsUnlimited {
		headers["X-RateLimit-Limit"] = "unlimited"
		headers["X-RateLimit-Remaining"] = "unlimited"
	} else {
		headers["X-RateLimit-Limit"] = fmt.Sprintf("%d", *stats.ItemsLimit)
		headers["X-RateLimit-Remaining"] = fmt.Sprintf("%d", *stats.ItemsRemaining)
	}

	headers["X-RateLimit-Reset"] = fmt.Sprintf("%d", stats.PeriodEnd.Unix())

	return headers
}

// Helper: buildUsageStats converts UsageTracking to UsageStats
func (s *usageService) buildUsageStats(usage *models.UsageTracking, tier models.SubscriptionTier) *models.UsageStats {
	stats := &models.UsageStats{
		ItemsProcessed: usage.ItemsProcessed,
		ItemsLimit:     usage.ItemsLimit,
		PeriodStart:    usage.PeriodStart,
		PeriodEnd:      usage.PeriodEnd,
		IsUnlimited:    tier.IsUnlimited(),
	}

	// Calculate remaining if there's a limit
	if usage.ItemsLimit != nil {
		remaining := *usage.ItemsLimit - usage.ItemsProcessed
		if remaining < 0 {
			remaining = 0
		}
		stats.ItemsRemaining = &remaining
	}

	return stats
}
