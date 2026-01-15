package repository

import (
	"context"
	"testing"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a valid test user with all required fields
func createTestUser(email, username string, tier models.SubscriptionTier) *models.User {
	return &models.User{
		ID:                     uuid.New(),
		Email:                  email,
		Username:               username,
		PasswordHash:           "$2a$10$validhashhere",
		SubscriptionTier:       tier,
		SubscriptionStatus:     "active", // Required by CHECK constraint
		MonthlyBgRemovalsUsed:  0,
		MonthlyBgRemovalsLimit: 50,
		MonthlyStorageUsedMB:   0,
		MonthlyStorageLimitMB:  500,
		FollowerCount:          0,
		FollowingCount:         0,
		IsVerified:             false,
	}
}

func TestUsageRepository_GetOrCreateCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	usageRepo := NewUsageRepository(db.Pool)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user with all required fields
	user := createTestUser("usage1@test.com", "usageuser1", models.TierFree)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Test: Create new usage record for free user
	t.Run("create_new_free_user", func(t *testing.T) {
		usage, err := usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

		require.NoError(t, err)
		require.NotNil(t, usage)
		assert.Equal(t, user.ID, usage.UserID)
		assert.Equal(t, 0, usage.ItemsProcessed)
		require.NotNil(t, usage.ItemsLimit)
		assert.Equal(t, 5, *usage.ItemsLimit)
	})

	// Test: Get existing record (should not create duplicate)
	t.Run("get_existing_record", func(t *testing.T) {
		usage2, err := usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

		require.NoError(t, err)
		require.NotNil(t, usage2)

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
	user := createTestUser("prouser@test.com", "prouser", models.TierPro)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("pro_user_unlimited", func(t *testing.T) {
		usage, err := usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierPro)

		require.NoError(t, err)
		require.NotNil(t, usage)
		assert.Nil(t, usage.ItemsLimit)
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

	user := createTestUser("increment@test.com", "incrementuser", models.TierFree)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	t.Run("increment_multiple_times", func(t *testing.T) {
		usage, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		assert.Equal(t, 0, usage.ItemsProcessed)

		for i := 1; i <= 3; i++ {
			err := usageRepo.IncrementUsage(ctx, user.ID)
			require.NoError(t, err)

			usage, _ = usageRepo.GetCurrentUsage(ctx, user.ID)
			assert.Equal(t, i, usage.ItemsProcessed)
		}
	})

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

	user := createTestUser("current@test.com", "currentuser", models.TierFree)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("no_record", func(t *testing.T) {
		usage, err := usageRepo.GetCurrentUsage(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, usage)
	})

	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

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

	user := createTestUser("update@test.com", "updateuser", models.TierFree)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	t.Run("update_to_unlimited", func(t *testing.T) {
		err := usageRepo.UpdateLimit(ctx, user.ID, nil)
		require.NoError(t, err)

		usage, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		assert.Nil(t, usage.ItemsLimit)
	})

	t.Run("update_to_limited", func(t *testing.T) {
		newLimit := 5
		err := usageRepo.UpdateLimit(ctx, user.ID, &newLimit)
		require.NoError(t, err)

		usage, _ := usageRepo.GetCurrentUsage(ctx, user.ID)
		require.NotNil(t, usage.ItemsLimit)
		assert.Equal(t, 5, *usage.ItemsLimit)
	})

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

	freeUser := createTestUser("free@reset.com", "freeuser", models.TierFree)
	err := userRepo.Create(ctx, freeUser)
	require.NoError(t, err)

	proUser := createTestUser("pro@reset.com", "proreset", models.TierPro)
	err = userRepo.Create(ctx, proUser)
	require.NoError(t, err)

	t.Run("reset_creates_records", func(t *testing.T) {
		count, err := usageRepo.ResetExpiredPeriods(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		freeUsage, _ := usageRepo.GetCurrentUsage(ctx, freeUser.ID)
		require.NotNil(t, freeUsage)
		require.NotNil(t, freeUsage.ItemsLimit)
		assert.Equal(t, 5, *freeUsage.ItemsLimit)

		proUsage, _ := usageRepo.GetCurrentUsage(ctx, proUser.ID)
		require.NotNil(t, proUsage)
		assert.Nil(t, proUsage.ItemsLimit)
	})

	t.Run("no_duplicates", func(t *testing.T) {
		count, err := usageRepo.ResetExpiredPeriods(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

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

	user := createTestUser("history@test.com", "historyuser", models.TierFree)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	_, _ = usageRepo.GetOrCreateCurrent(ctx, user.ID, models.TierFree)

	t.Run("get_history", func(t *testing.T) {
		history, err := usageRepo.GetUsageHistory(ctx, user.ID, 12)
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, user.ID, history[0].UserID)
	})

	t.Run("no_history", func(t *testing.T) {
		nonExistentUser := uuid.New()
		history, err := usageRepo.GetUsageHistory(ctx, nonExistentUser, 12)
		require.NoError(t, err)
		assert.Len(t, history, 0)
	})

	t.Run("limit_works", func(t *testing.T) {
		history, err := usageRepo.GetUsageHistory(ctx, user.ID, 1)
		require.NoError(t, err)
		assert.Len(t, history, 1)
	})
}

func TestGetCurrentPeriod(t *testing.T) {
	t.Run("current_period", func(t *testing.T) {
		start, end := GetCurrentPeriod()

		assert.Equal(t, 1, start.Day())
		assert.Equal(t, 0, start.Hour())
		assert.Equal(t, 0, start.Minute())
		assert.Equal(t, 0, start.Second())

		assert.True(t, end.After(start))
		assert.Equal(t, start.Month(), end.Month())
		assert.Equal(t, 23, end.Hour())
		assert.Equal(t, 59, end.Minute())
		assert.Equal(t, 59, end.Second())
	})
}
