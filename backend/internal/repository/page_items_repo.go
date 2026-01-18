package repository

import (
	"context"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PageItemsRepository interface {
	Create(ctx context.Context, userID uuid.UUID, item *models.PageItem) error
	ListByPage(ctx context.Context, userID, pageID uuid.UUID) ([]*models.PageItem, error)
	Update(ctx context.Context, userID uuid.UUID, update *models.PageItemUpdate) (*models.PageItem, error)
	Delete(ctx context.Context, userID, pageID, pageItemID uuid.UUID) error
}

type pageItemsRepository struct {
	db *database.DB
}

func NewPageItemsRepository(db *database.DB) PageItemsRepository {
	return &pageItemsRepository{db: db}
}

func (r *pageItemsRepository) Create(ctx context.Context, userID uuid.UUID, item *models.PageItem) error {
	query := `
		INSERT INTO content.page_items (
			id,
			page_id,
			item_id,
			position_x,
			position_y,
			width,
			height,
			rotation,
			z_index,
			opacity,
			filters
		)
		SELECT
			$1,
			pg.id,
			it.id,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11
		FROM content.pages pg
		JOIN content.projects pr ON pg.project_id = pr.id
		JOIN content.items it ON it.id = $3 AND it.user_id = $12
		WHERE pg.id = $2 AND pr.user_id = $12
		RETURNING created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		item.ID,
		item.PageID,
		item.ItemID,
		item.PositionX,
		item.PositionY,
		item.Width,
		item.Height,
		item.Rotation,
		item.ZIndex,
		item.Opacity,
		jsonRawOrNil(item.Filters),
		userID,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.ErrNotFound("Page or Item")
		}
		return utils.ErrDatabase("Failed to create page item", err)
	}

	return nil
}

func (r *pageItemsRepository) ListByPage(ctx context.Context, userID, pageID uuid.UUID) ([]*models.PageItem, error) {
	query := `
		SELECT
			pi.id,
			pi.page_id,
			pi.item_id,
			pi.position_x,
			pi.position_y,
			pi.width,
			pi.height,
			pi.rotation,
			pi.z_index,
			pi.opacity,
			pi.filters,
			pi.created_at,
			pi.updated_at
		FROM content.page_items pi
		JOIN content.pages pg ON pi.page_id = pg.id
		JOIN content.projects pr ON pg.project_id = pr.id
		WHERE pr.user_id = $1 AND pg.id = $2
		ORDER BY pi.z_index ASC, pi.created_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, pageID)
	if err != nil {
		return nil, utils.ErrDatabase("Failed to list page items", err)
	}
	defer rows.Close()

	items := []*models.PageItem{}
	for rows.Next() {
		item := &models.PageItem{}
		if err := rows.Scan(
			&item.ID,
			&item.PageID,
			&item.ItemID,
			&item.PositionX,
			&item.PositionY,
			&item.Width,
			&item.Height,
			&item.Rotation,
			&item.ZIndex,
			&item.Opacity,
			&item.Filters,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, utils.ErrDatabase("Failed to scan page items", err)
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, utils.ErrDatabase("Failed to iterate page items", rows.Err())
	}

	return items, nil
}

func (r *pageItemsRepository) Update(ctx context.Context, userID uuid.UUID, update *models.PageItemUpdate) (*models.PageItem, error) {
	query := `
		UPDATE content.page_items
		SET
			position_x = COALESCE($3, position_x),
			position_y = COALESCE($4, position_y),
			width = COALESCE($5, width),
			height = COALESCE($6, height),
			rotation = COALESCE($7, rotation),
			z_index = COALESCE($8, z_index),
			opacity = COALESCE($9, opacity),
			filters = COALESCE($10, filters),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
			AND page_id = $2
			AND page_id IN (
				SELECT pg.id
				FROM content.pages pg
				JOIN content.projects pr ON pg.project_id = pr.id
				WHERE pr.user_id = $11
			)
		RETURNING
			id,
			page_id,
			item_id,
			position_x,
			position_y,
			width,
			height,
			rotation,
			z_index,
			opacity,
			filters,
			created_at,
			updated_at
	`

	item := &models.PageItem{}
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		update.ID,
		update.PageID,
		update.PositionX,
		update.PositionY,
		update.Width,
		update.Height,
		update.Rotation,
		update.ZIndex,
		update.Opacity,
		jsonRawOrNil(update.Filters),
		userID,
	).Scan(
		&item.ID,
		&item.PageID,
		&item.ItemID,
		&item.PositionX,
		&item.PositionY,
		&item.Width,
		&item.Height,
		&item.Rotation,
		&item.ZIndex,
		&item.Opacity,
		&item.Filters,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Page item")
		}
		return nil, utils.ErrDatabase("Failed to update page item", err)
	}

	return item, nil
}

func (r *pageItemsRepository) Delete(ctx context.Context, userID, pageID, pageItemID uuid.UUID) error {
	query := `
		DELETE FROM content.page_items
		WHERE id = $1
			AND page_id = $2
			AND page_id IN (
				SELECT pg.id
				FROM content.pages pg
				JOIN content.projects pr ON pg.project_id = pr.id
				WHERE pr.user_id = $3
			)
	`

	result, err := r.db.Pool.Exec(ctx, query, pageItemID, pageID, userID)
	if err != nil {
		return utils.ErrDatabase("Failed to delete page item", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrNotFound("Page item")
	}

	return nil
}
