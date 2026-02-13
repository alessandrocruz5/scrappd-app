package repository

import (
	"context"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ItemsRepository interface {
	Create(ctx context.Context, item *models.Item) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Item, int, error)
	GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error)
	SoftDelete(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error)
	UpdateProcessingStarted(ctx context.Context, userID, itemID uuid.UUID, startedAt time.Time) error
	UpdateProcessingCompleted(ctx context.Context, userID, itemID uuid.UUID, processedKey, processedURL string, processedSize int64, modelVersion string, completedAt time.Time) (bool, error)
	UpdateProcessingFailed(ctx context.Context, userID, itemID uuid.UUID, errorMsg string, completedAt time.Time) error
	UpdateProcessingCancelled(ctx context.Context, userID, itemID uuid.UUID, completedAt time.Time) (bool, error)
}

type itemsRepository struct {
	db *database.DB
}

func NewItemsRepository(db *database.DB) ItemsRepository {
	return &itemsRepository{db: db}
}

func (r *itemsRepository) Create(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO content.items (
			id, user_id,
			original_image_key, original_image_url, original_file_size_bytes,
			original_width, original_height,
			processed_image_key, processed_image_url, processed_file_size_bytes,
			processing_status, ml_model_version, processing_started_at, processing_completed_at, processing_error,
			mime_type, item_name, item_category, tags
		) VALUES (
			$1, $2,
			$3, $4, $5,
			$6, $7,
			$8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19
		)
		RETURNING created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		item.ID,
		item.UserID,
		item.OriginalImageKey,
		item.OriginalImageURL,
		item.OriginalFileSize,
		item.OriginalWidth,
		item.OriginalHeight,
		item.ProcessedImageKey,
		item.ProcessedImageURL,
		item.ProcessedFileSize,
		item.ProcessingStatus,
		item.MLModelVersion,
		item.ProcessingStartedAt,
		item.ProcessingCompletedAt,
		item.ProcessingError,
		item.MimeType,
		item.ItemName,
		item.ItemCategory,
		item.Tags,
	).Scan(&item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return utils.ErrDatabase("Failed to create item", err)
	}

	return nil
}

func (r *itemsRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Item, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	countQuery := `
		SELECT COUNT(*)
		FROM content.items
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var total int
	if err := r.db.Pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, utils.ErrDatabase("Failed to count items", err)
	}

	query := `
		SELECT
			id, user_id,
			original_image_key, original_image_url, original_file_size_bytes,
			original_width, original_height,
			processed_image_key, processed_image_url, processed_file_size_bytes,
			processing_status, ml_model_version, processing_started_at, processing_completed_at, processing_error,
			mime_type, item_name, item_category, tags,
			created_at, updated_at, deleted_at
		FROM content.items
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrDatabase("Failed to list items", err)
	}
	defer rows.Close()

	items := []*models.Item{}
	for rows.Next() {
		item := &models.Item{}
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.OriginalImageKey,
			&item.OriginalImageURL,
			&item.OriginalFileSize,
			&item.OriginalWidth,
			&item.OriginalHeight,
			&item.ProcessedImageKey,
			&item.ProcessedImageURL,
			&item.ProcessedFileSize,
			&item.ProcessingStatus,
			&item.MLModelVersion,
			&item.ProcessingStartedAt,
			&item.ProcessingCompletedAt,
			&item.ProcessingError,
			&item.MimeType,
			&item.ItemName,
			&item.ItemCategory,
			&item.Tags,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, 0, utils.ErrDatabase("Failed to scan items", err)
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, 0, utils.ErrDatabase("Failed to iterate items", rows.Err())
	}

	return items, total, nil
}

func (r *itemsRepository) GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	query := `
		SELECT
			id, user_id,
			original_image_key, original_image_url, original_file_size_bytes,
			original_width, original_height,
			processed_image_key, processed_image_url, processed_file_size_bytes,
			processing_status, ml_model_version, processing_started_at, processing_completed_at, processing_error,
			mime_type, item_name, item_category, tags,
			created_at, updated_at, deleted_at
		FROM content.items
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	item := &models.Item{}
	err := r.db.Pool.QueryRow(ctx, query, itemID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.OriginalImageKey,
		&item.OriginalImageURL,
		&item.OriginalFileSize,
		&item.OriginalWidth,
		&item.OriginalHeight,
		&item.ProcessedImageKey,
		&item.ProcessedImageURL,
		&item.ProcessedFileSize,
		&item.ProcessingStatus,
		&item.MLModelVersion,
		&item.ProcessingStartedAt,
		&item.ProcessingCompletedAt,
		&item.ProcessingError,
		&item.MimeType,
		&item.ItemName,
		&item.ItemCategory,
		&item.Tags,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Item")
		}
		return nil, utils.ErrDatabase("Failed to get item", err)
	}

	return item, nil
}

func (r *itemsRepository) SoftDelete(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	query := `
		UPDATE content.items
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING
			id, user_id,
			original_image_key, original_image_url,
			processed_image_key, processed_image_url,
			created_at, updated_at, deleted_at
	`

	item := &models.Item{}
	err := r.db.Pool.QueryRow(ctx, query, itemID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.OriginalImageKey,
		&item.OriginalImageURL,
		&item.ProcessedImageKey,
		&item.ProcessedImageURL,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Item")
		}
		return nil, utils.ErrDatabase("Failed to delete item", err)
	}

	return item, nil
}

func (r *itemsRepository) UpdateProcessingStarted(ctx context.Context, userID, itemID uuid.UUID, startedAt time.Time) error {
	query := `
		UPDATE content.items
		SET processing_status = 'processing',
			processing_started_at = $3,
			processing_error = NULL
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Pool.Exec(ctx, query, itemID, userID, startedAt)
	if err != nil {
		return utils.ErrDatabase("Failed to update processing status", err)
	}

	return nil
}

func (r *itemsRepository) UpdateProcessingCompleted(ctx context.Context, userID, itemID uuid.UUID, processedKey, processedURL string, processedSize int64, modelVersion string, completedAt time.Time) (bool, error) {
	query := `
		UPDATE content.items
		SET processing_status = 'completed',
			processed_image_key = $3,
			processed_image_url = $4,
			processed_file_size_bytes = $5,
			ml_model_version = $6,
			processing_completed_at = $7,
			processing_error = NULL
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
			AND processing_status = 'processing'
	`

	tag, err := r.db.Pool.Exec(ctx, query, itemID, userID, processedKey, processedURL, processedSize, modelVersion, completedAt)
	if err != nil {
		return false, utils.ErrDatabase("Failed to update processing result", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (r *itemsRepository) UpdateProcessingFailed(ctx context.Context, userID, itemID uuid.UUID, errorMsg string, completedAt time.Time) error {
	query := `
		UPDATE content.items
		SET processing_status = 'failed',
			processing_error = $3,
			processing_completed_at = $4
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Pool.Exec(ctx, query, itemID, userID, errorMsg, completedAt)
	if err != nil {
		return utils.ErrDatabase("Failed to update processing failure", err)
	}

	return nil
}

func (r *itemsRepository) UpdateProcessingCancelled(ctx context.Context, userID, itemID uuid.UUID, completedAt time.Time) (bool, error) {
	query := `
		UPDATE content.items
		SET processing_status = 'failed',
			processing_error = 'cancelled',
			processing_completed_at = $3
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
			AND processing_status IN ('pending', 'processing')
	`

	tag, err := r.db.Pool.Exec(ctx, query, itemID, userID, completedAt)
	if err != nil {
		return false, utils.ErrDatabase("Failed to cancel processing", err)
	}

	return tag.RowsAffected() > 0, nil
}
