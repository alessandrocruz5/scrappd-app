package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUsageRepository is a mock implementation of UsageRepository
type MockUsageRepository struct {
	mock.Mock
}

func (m *MockUsageRepository) GetOrCreateCurrent(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (*models.UsageTracking, error) {
	args := m.Called(ctx, userID, tier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UsageTracking), args.Error(1)
}

func (m *MockUsageRepository) IncrementUsage(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUsageRepository) GetCurrentUsage(ctx context.Context, userID uuid.UUID) (*models.UsageTracking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UsageTracking), args.Error(1)
}

func (m *MockUsageRepository) UpdateLimit(ctx context.Context, userID uuid.UUID, limit *int) error {
	args := m.Called(ctx, userID, limit)
	return args.Error(0)
}

func (m *MockUsageRepository) ResetExpiredPeriods(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockUsageRepository) GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UsageTracking), args.Error(1)
}

func TestUsageService_CheckAndIncrementUsage_FreeUser(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Test: Free user with quota remaining
	t.Run("free_user_has_quota", func(t *testing.T) {
		limit := 5
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			ItemsProcessed: 2,
			ItemsLimit:     &limit,
		}

		mockRepo.On("GetOrCreateCurrent", ctx, userID, models.TierFree).Return(usage, nil).Once()
		mockRepo.On("IncrementUsage", ctx, userID).Return(nil).Once()

		canProcess, stats, err := service.CheckAndIncrementUsage(ctx, userID, models.TierFree)

		require.NoError(t, err)
		assert.True(t, canProcess)
		require.NotNil(t, stats)
		assert.Equal(t, 3, stats.ItemsProcessed) // Should show incremented value
		assert.Equal(t, 5, *stats.ItemsLimit)
		assert.Equal(t, 2, *stats.ItemsRemaining) // 5 - 3 = 2

		mockRepo.AssertExpectations(t)
	})

	// Test: Free user at limit
	t.Run("free_user_at_limit", func(t *testing.T) {
		limit := 5
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			ItemsProcessed: 5,
			ItemsLimit:     &limit,
		}

		mockRepo.On("GetOrCreateCurrent", ctx, userID, models.TierFree).Return(usage, nil).Once()

		canProcess, stats, err := service.CheckAndIncrementUsage(ctx, userID, models.TierFree)

		require.NoError(t, err)
		assert.False(t, canProcess)
		require.NotNil(t, stats)
		assert.Equal(t, 5, stats.ItemsProcessed)
		assert.Equal(t, 0, *stats.ItemsRemaining)

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "IncrementUsage")
	})

	// Test: Free user over limit
	t.Run("free_user_over_limit", func(t *testing.T) {
		limit := 5
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			ItemsProcessed: 6,
			ItemsLimit:     &limit,
		}

		mockRepo.On("GetOrCreateCurrent", ctx, userID, models.TierFree).Return(usage, nil).Once()

		canProcess, stats, err := service.CheckAndIncrementUsage(ctx, userID, models.TierFree)

		require.NoError(t, err)
		assert.False(t, canProcess)
		assert.Equal(t, 0, *stats.ItemsRemaining)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_CheckAndIncrementUsage_ProUser(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Test: Pro user has unlimited
	t.Run("pro_user_unlimited", func(t *testing.T) {
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			ItemsProcessed: 100,
			ItemsLimit:     nil, // Unlimited
		}

		mockRepo.On("GetOrCreateCurrent", ctx, userID, models.TierPro).Return(usage, nil).Once()
		mockRepo.On("IncrementUsage", ctx, userID).Return(nil).Once()

		canProcess, stats, err := service.CheckAndIncrementUsage(ctx, userID, models.TierPro)

		require.NoError(t, err)
		assert.True(t, canProcess)
		require.NotNil(t, stats)
		assert.True(t, stats.IsUnlimited)
		assert.Nil(t, stats.ItemsLimit)
		assert.Nil(t, stats.ItemsRemaining)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_GetCurrentUsageStats(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Test: Get stats for free user
	t.Run("free_user_stats", func(t *testing.T) {
		limit := 5
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			ItemsProcessed: 3,
			ItemsLimit:     &limit,
		}

		mockRepo.On("GetCurrentUsage", ctx, userID).Return(usage, nil).Once()

		stats, err := service.GetCurrentUsageStats(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 3, stats.ItemsProcessed)
		assert.Equal(t, 5, *stats.ItemsLimit)
		assert.Equal(t, 2, *stats.ItemsRemaining)
		assert.False(t, stats.IsUnlimited)

		mockRepo.AssertExpectations(t)
	})

	// Test: Get stats for pro user
	t.Run("pro_user_stats", func(t *testing.T) {
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			ItemsProcessed: 50,
			ItemsLimit:     nil,
		}

		mockRepo.On("GetCurrentUsage", ctx, userID).Return(usage, nil).Once()

		stats, err := service.GetCurrentUsageStats(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 50, stats.ItemsProcessed)
		assert.Nil(t, stats.ItemsLimit)
		assert.Nil(t, stats.ItemsRemaining)
		assert.True(t, stats.IsUnlimited)

		mockRepo.AssertExpectations(t)
	})

	// Test: No usage record found
	t.Run("no_usage_record", func(t *testing.T) {
		mockRepo.On("GetCurrentUsage", ctx, userID).Return(nil, nil).Once()

		stats, err := service.GetCurrentUsageStats(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, stats)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_GetUsageHistory(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Test: Get history with records
	t.Run("get_history", func(t *testing.T) {
		limit := 5
		history := []models.UsageTracking{
			{
				ID:             uuid.New(),
				UserID:         userID,
				ItemsProcessed: 3,
				ItemsLimit:     &limit,
			},
			{
				ID:             uuid.New(),
				UserID:         userID,
				ItemsProcessed: 5,
				ItemsLimit:     &limit,
			},
		}

		mockRepo.On("GetUsageHistory", ctx, userID, 12).Return(history, nil).Once()

		result, err := service.GetUsageHistory(ctx, userID, 12)

		require.NoError(t, err)
		assert.Len(t, result, 2)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_UpdateSubscriptionTier(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Test: Upgrade to Pro
	t.Run("upgrade_to_pro", func(t *testing.T) {
		var nilLimit *int = nil
		mockRepo.On("UpdateLimit", ctx, userID, nilLimit).Return(nil).Once()

		err := service.UpdateSubscriptionTier(ctx, userID, models.TierPro)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Test: Downgrade to Free
	t.Run("downgrade_to_free", func(t *testing.T) {
		limit := 5
		mockRepo.On("UpdateLimit", ctx, userID, &limit).Return(nil).Once()

		err := service.UpdateSubscriptionTier(ctx, userID, models.TierFree)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_ResetExpiredPeriods(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()

	// Test: Reset returns count
	t.Run("reset_success", func(t *testing.T) {
		mockRepo.On("ResetExpiredPeriods", ctx).Return(10, nil).Once()

		count, err := service.ResetExpiredPeriods(ctx)

		require.NoError(t, err)
		assert.Equal(t, 10, count)

		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_InitializeUsageForNewUser(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Test: Initialize for free user
	t.Run("initialize_free_user", func(t *testing.T) {
		limit := 5
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			ItemsProcessed: 0,
			ItemsLimit:     &limit,
		}

		mockRepo.On("GetOrCreateCurrent", ctx, userID, models.TierFree).Return(usage, nil).Once()

		err := service.InitializeUsageForNewUser(ctx, userID, models.TierFree)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Test: Initialize for pro user
	t.Run("initialize_pro_user", func(t *testing.T) {
		usage := &models.UsageTracking{
			ID:             uuid.New(),
			UserID:         userID,
			ItemsProcessed: 0,
			ItemsLimit:     nil,
		}

		mockRepo.On("GetOrCreateCurrent", ctx, userID, models.TierPro).Return(usage, nil).Once()

		err := service.InitializeUsageForNewUser(ctx, userID, models.TierPro)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUsageService_GetRateLimitHeaders(t *testing.T) {
	service := NewUsageService(nil) // No mock needed for this test

	// Test: Free user headers
	t.Run("free_user_headers", func(t *testing.T) {
		limit := 5
		remaining := 3
		stats := &models.UsageStats{
			ItemsProcessed: 2,
			ItemsLimit:     &limit,
			ItemsRemaining: &remaining,
			PeriodEnd:      time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			IsUnlimited:    false,
		}

		headers := service.GetRateLimitHeaders(stats)

		assert.Equal(t, "5", headers["X-RateLimit-Limit"])
		assert.Equal(t, "3", headers["X-RateLimit-Remaining"])
		assert.Equal(t, fmt.Sprintf("%d", stats.PeriodEnd.Unix()), headers["X-RateLimit-Reset"])
	})

	// Test: Pro user headers
	t.Run("pro_user_headers", func(t *testing.T) {
		stats := &models.UsageStats{
			ItemsProcessed: 50,
			ItemsLimit:     nil,
			ItemsRemaining: nil,
			PeriodEnd:      time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			IsUnlimited:    true,
		}

		headers := service.GetRateLimitHeaders(stats)

		assert.Equal(t, "unlimited", headers["X-RateLimit-Limit"])
		assert.Equal(t, "unlimited", headers["X-RateLimit-Remaining"])
		assert.Equal(t, fmt.Sprintf("%d", stats.PeriodEnd.Unix()), headers["X-RateLimit-Reset"])
	})
}
