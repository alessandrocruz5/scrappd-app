package repository

import (
	"context"
	"testing"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageRepository_GetOrCreateCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &models.User{
		ID:               uuid.New(),
		Email:            "usage1@test.com",
		Username:         "usageuser1",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierFree,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Test: Create new usage record for free user
	t.Run("create_new_free_user", func(t *testing.T) {
		usage, err := usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

		require.NoError(t, err)
		require.NotNil(t, usage)
		assert.Equal(t, user.ID, usage.UserID)
		assert.Equal(t, 0, usage.ItemsProcessed)
		assert.NotNil(t, usage.ItemsLimit)
		assert.Equal(t, 5, *usage.ItemsLimit)
	})

	// Test: Get existing record (should not create duplicate)
	t.Run("get_existing_record", func(t *testing.T) {
		// Call again - should return existing record
		usage2, err := usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

		require.NoError(t, err)
		require.NotNil(t, usage2)

		// Should be the same record
		usage1, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		assert.Equal(t, usage1.ID, usage2.ID)
	})
}

func TestUsageRepository_GetOrCreateCurrent_ProUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create pro user
	user := &models.User{
		ID:               uuid.New(),
		Email:            "prouser@test.com",
		Username:         "prouser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierPro,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Test: Pro user gets unlimited
	t.Run("pro_user_unlimited", func(t *testing.T) {
		usage, err := usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierPro)

		require.NoError(t, err)
		require.NotNil(t, usage)
		assert.Nil(t, usage.ItemsLimit) // Pro users have no limit
	})
}

func TestUsageRepository_IncrementUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:               uuid.New(),
		Email:            "increment@test.com",
		Username:         "incrementuser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierFree,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	// Test: Increment counter multiple times
	t.Run("increment_multiple_times", func(t *testing.T) {
		// Initial state
		usage, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		assert.Equal(t, 0, usage.ItemsProcessed)

		// Increment 3 times
		for i := 1; i <= 3; i++ {
			err := usageRepo.IncrementUsage(ctx, user.ID)
			require.NoError(t, err)

			usage, _ = usageRepo.GetCurrentUsage(ctx, user.ID)
			assert.Equal(t, i, usage.ItemsProcessed)
		}
	})

	// Test: Error when no record exists
	t.Run("error_no_record", func(t *testing.T) {
		nonExistentUser := uuid.New()
		err := usageRepo.IncrementUsage(ctx, nonExistentUser)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no usage record found")
	})
}

func TestUsageRepository_GetCurrentUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:               uuid.New(),
		Email:            "current@test.com",
		Username:         "currentuser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierFree,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Test: No record exists initially
	t.Run("no_record", func(t *testing.T) {
		usage, err := usageRepo.GetCurrentUsage(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, usage)
	})

	// Create record
	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	// Test: Get existing record
	t.Run("get_existing", func(t *testing.T) {
		usage, err := usageRepo.GetCurrentUsage(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, usage)
		assert.Equal(t, user.ID, usage.UserID)
		assert.NotZero(t, usage.PeriodStart)
		assert.NotZero(t, usage.PeriodEnd)
	})
}

func TestUsageRepository_UpdateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:               uuid.New(),
		Email:            "update@test.com",
		Username:         "updateuser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierFree,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	// Test: Update to unlimited (upgrade to pro)
	t.Run("update_to_unlimited", func(t *testing.T) {
		err := usageRepo.UpdateLimit(ctx, user.ID, nil)
		require.NoError(t, err)

		usage, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		assert.Nil(t, usage.ItemsLimit)
	})

	// Test: Update to limited (downgrade to free)
	t.Run("update_to_limited", func(t *testing.T) {
		newLimit := 5
		err := usageRepo.UpdateLimit(ctx, user.ID, &newLimit)
		require.NoError(t, err)

		usage, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		require.NotNil(t, usage.ItemsLimit)
		assert.Equal(t, 5, *usage.ItemsLimit)
	})

	// Test: Error when no record exists
	t.Run("error_no_record", func(t *testing.T) {
		nonExistentUser := uuid.New()
		newLimit := 10
		err := usageRepo.UpdateLimit(ctx, nonExistentUser, &newLimit)
		require.Error(t, err)
	})
}

func TestUsageRepository_ResetExpiredPeriods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create multiple users with different tiers
	freeUser := &models.User{
		ID:               uuid.New(),
		Email:            "free@reset.com",
		Username:         "freeuser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierFree,
	}
	err := userRepo.Create(ctx, freeUser)
	require.NoError(t, err)

	proUser := &models.User{
		ID:               uuid.New(),
		Email:            "pro@reset.com",
		Username:         "prouser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierPro,
	}
	err = userRepo.Create(ctx, proUser)
	require.NoError(t, err)

	// Test: Reset creates records for users without current period
	t.Run("reset_creates_records", func(t *testing.T) {
		count, err := usageRepo.ResetExpiredPeriods(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count) // Should create for both users

		// Verify free user has limit
		freeUsage, _ := usageRepo.GetCurrentUsage(ctx, freeUser.ID)
		require.NotNil(t, freeUsage)
		require.NotNil(t, freeUsage.ItemsLimit)
		assert.Equal(t, 5, *freeUsage.ItemsLimit)

		// Verify pro user has no limit
		proUsage, _ := usageRepo.GetCurrentUsage(ctx, proUser.ID)
		require.NotNil(t, proUsage)
		assert.Nil(t, proUsage.ItemsLimit)
	})

	// Test: Reset doesn't create duplicates
	t.Run("no_duplicates", func(t *testing.T) {
		count, err := usageRepo.ResetExpiredPeriods(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count) // No new records created

		// Verify still only one record per user
		freeUsage, _ := usageRepo.GetCurrentUsage(ctx, freeUser.ID)
		require.NotNil(t, freeUsage)
	})
}

func TestUsageRepository_GetUsageHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:               uuid.New(),
		Email:            "history@test.com",
		Username:         "historyuser",
		PasswordHash:     "hash",
		SubscriptionTier: models.TierFree,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create current period record
	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	// Test: Get history with records
	t.Run("get_history", func(t *testing.T) {
		history, err := usageRepo.GetUsageHistory(ctx, user.ID, 12)
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, user.ID, history[0].UserID)
	})

	// Test: Get history for user with no records
	t.Run("no_history", func(t *testing.T) {
		nonExistentUser := uuid.New()
		history, err := usageRepo.GetUsageHistory(ctx, nonExistentUser, 12)
		require.NoError(t, err)
		assert.Len(t, history, 0)
	})

	// Test: Limit works correctly
	t.Run("limit_works", func(t *testing.T) {
		history, err := usageRepo.GetUsageHistory(ctx, user.ID, 1)
		require.NoError(t, err)
		assert.Len(t, history, 1)
	})
}

func TestGetCurrentPeriod(t *testing.T) {
	// Test: Current period is correctly calculated
	t.Run("current_period", func(t *testing.T) {
		start, end := GetCurrentPeriod()

		// Start should be first day of current month at midnight
		assert.Equal(t, 1, start.Day())
		assert.Equal(t, 0, start.Hour())
		assert.Equal(t, 0, start.Minute())
		assert.Equal(t, 0, start.Second())

		// End should be last second of current month
		assert.True(t, end.After(start))
		assert.Equal(t, start.Month(), end.Month())
		assert.Equal(t, 23, end.Hour())
		assert.Equal(t, 59, end.Minute())
		assert.Equal(t, 59, end.Second())
	})
}
