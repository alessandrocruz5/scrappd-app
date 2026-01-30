package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemsRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	itemsRepo := NewItemsRepository(db)

	user := createTestUser("items@test.com", "itemsuser", models.TierFree)
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	item := &models.Item{
		ID:                uuid.New(),
		UserID:            user.ID,
		OriginalImageKey:  "uploads/original.png",
		OriginalImageURL:  "signed-original",
		ProcessedImageKey: strPtr("uploads/processed.png"),
		ProcessedImageURL: strPtr("signed-processed"),
		ProcessingStatus:  "completed",
		MLModelVersion:    strPtr("birefnet-general"),
		MimeType:          strPtr("image/png"),
		Tags:              []string{"tag1", "tag2"},
		ProcessingStartedAt:  timePtr(time.Now().Add(-1 * time.Minute)),
		ProcessingCompletedAt: timePtr(time.Now()),
	}

	err = itemsRepo.Create(ctx, item)
	require.NoError(t, err)

	t.Run("get_item", func(t *testing.T) {
		got, err := itemsRepo.GetByID(ctx, user.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, got.ID)
		assert.Equal(t, item.OriginalImageKey, got.OriginalImageKey)
		assert.Equal(t, item.Tags, got.Tags)
	})

	t.Run("list_items", func(t *testing.T) {
		items, total, err := itemsRepo.ListByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, items, 1)
	})

	t.Run("soft_delete", func(t *testing.T) {
		deleted, err := itemsRepo.SoftDelete(ctx, user.ID, item.ID)
		require.NoError(t, err)
		assert.NotNil(t, deleted.DeletedAt)

		_, err = itemsRepo.GetByID(ctx, user.ID, item.ID)
		assert.Error(t, err)
	})
}

func timePtr(val time.Time) *time.Time {
	return &val
}

func strPtr(val string) *string {
	return &val
}
