package repository

import (
	"context"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PagesRepository interface {
	Create(ctx context.Context, userID uuid.UUID, page *models.Page) error
	ListByProject(ctx context.Context, userID, projectID uuid.UUID, limit, offset int) ([]*models.Page, int, error)
	GetByID(ctx context.Context, userID, pageID uuid.UUID) (*models.Page, error)
	Update(ctx context.Context, userID uuid.UUID, update *models.PageUpdate) (*models.Page, error)
	Delete(ctx context.Context, userID, pageID uuid.UUID) error
}

type pagesRepository struct {
	db *database.DB
}

func NewPagesRepository(db *database.DB) PagesRepository {
	return &pagesRepository{db: db}
}

func (r *pagesRepository) Create(ctx context.Context, userID uuid.UUID, page *models.Page) error {
	query := `
		INSERT INTO content.pages (
			id,
			project_id,
			page_number,
			title,
			canvas_width,
			canvas_height,
			background_color,
			background_image_url,
			background_pattern,
			layout_template
		)
		SELECT
			$1,
			p.id,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9
		FROM content.projects p
		WHERE p.id = $2 AND p.user_id = $10
		RETURNING created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		page.ID,
		page.ProjectID,
		page.PageNumber,
		page.Title,
		page.CanvasWidth,
		page.CanvasHeight,
		page.BackgroundColor,
		page.BackgroundImageURL,
		page.BackgroundPattern,
		page.LayoutTemplate,
		userID,
	).Scan(&page.CreatedAt, &page.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.ErrNotFound("Project")
		}
		return utils.ErrDatabase("Failed to create page", err)
	}

	return nil
}

func (r *pagesRepository) ListByProject(ctx context.Context, userID, projectID uuid.UUID, limit, offset int) ([]*models.Page, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	countQuery := `
		SELECT COUNT(*)
		FROM content.pages pg
		JOIN content.projects pr ON pg.project_id = pr.id
		WHERE pr.user_id = $1 AND pg.project_id = $2
	`

	var total int
	if err := r.db.Pool.QueryRow(ctx, countQuery, userID, projectID).Scan(&total); err != nil {
		return nil, 0, utils.ErrDatabase("Failed to count pages", err)
	}

	query := `
		SELECT
			pg.id,
			pg.project_id,
			pg.page_number,
			pg.title,
			pg.canvas_width,
			pg.canvas_height,
			pg.background_color,
			pg.background_image_url,
			pg.background_pattern,
			pg.layout_template,
			pg.created_at,
			pg.updated_at
		FROM content.pages pg
		JOIN content.projects pr ON pg.project_id = pr.id
		WHERE pr.user_id = $1 AND pg.project_id = $2
		ORDER BY pg.page_number ASC, pg.created_at ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, projectID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrDatabase("Failed to list pages", err)
	}
	defer rows.Close()

	pages := []*models.Page{}
	for rows.Next() {
		page := &models.Page{}
		if err := rows.Scan(
			&page.ID,
			&page.ProjectID,
			&page.PageNumber,
			&page.Title,
			&page.CanvasWidth,
			&page.CanvasHeight,
			&page.BackgroundColor,
			&page.BackgroundImageURL,
			&page.BackgroundPattern,
			&page.LayoutTemplate,
			&page.CreatedAt,
			&page.UpdatedAt,
		); err != nil {
			return nil, 0, utils.ErrDatabase("Failed to scan pages", err)
		}
		pages = append(pages, page)
	}

	if rows.Err() != nil {
		return nil, 0, utils.ErrDatabase("Failed to iterate pages", rows.Err())
	}

	return pages, total, nil
}

func (r *pagesRepository) GetByID(ctx context.Context, userID, pageID uuid.UUID) (*models.Page, error) {
	query := `
		SELECT
			pg.id,
			pg.project_id,
			pg.page_number,
			pg.title,
			pg.canvas_width,
			pg.canvas_height,
			pg.background_color,
			pg.background_image_url,
			pg.background_pattern,
			pg.layout_template,
			pg.created_at,
			pg.updated_at
		FROM content.pages pg
		JOIN content.projects pr ON pg.project_id = pr.id
		WHERE pr.user_id = $1 AND pg.id = $2
	`

	page := &models.Page{}
	err := r.db.Pool.QueryRow(ctx, query, userID, pageID).Scan(
		&page.ID,
		&page.ProjectID,
		&page.PageNumber,
		&page.Title,
		&page.CanvasWidth,
		&page.CanvasHeight,
		&page.BackgroundColor,
		&page.BackgroundImageURL,
		&page.BackgroundPattern,
		&page.LayoutTemplate,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Page")
		}
		return nil, utils.ErrDatabase("Failed to get page", err)
	}

	return page, nil
}

func (r *pagesRepository) Update(ctx context.Context, userID uuid.UUID, update *models.PageUpdate) (*models.Page, error) {
	query := `
		UPDATE content.pages
		SET
			page_number = COALESCE($2, page_number),
			title = COALESCE($3, title),
			canvas_width = COALESCE($4, canvas_width),
			canvas_height = COALESCE($5, canvas_height),
			background_color = COALESCE($6, background_color),
			background_image_url = COALESCE($7, background_image_url),
			background_pattern = COALESCE($8, background_pattern),
			layout_template = COALESCE($9, layout_template),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
			AND project_id IN (SELECT id FROM content.projects WHERE user_id = $10)
		RETURNING
			id,
			project_id,
			page_number,
			title,
			canvas_width,
			canvas_height,
			background_color,
			background_image_url,
			background_pattern,
			layout_template,
			created_at,
			updated_at
	`

	page := &models.Page{}
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		update.ID,
		update.PageNumber,
		update.Title,
		update.CanvasWidth,
		update.CanvasHeight,
		update.BackgroundColor,
		update.BackgroundImageURL,
		update.BackgroundPattern,
		update.LayoutTemplate,
		userID,
	).Scan(
		&page.ID,
		&page.ProjectID,
		&page.PageNumber,
		&page.Title,
		&page.CanvasWidth,
		&page.CanvasHeight,
		&page.BackgroundColor,
		&page.BackgroundImageURL,
		&page.BackgroundPattern,
		&page.LayoutTemplate,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Page")
		}
		return nil, utils.ErrDatabase("Failed to update page", err)
	}

	return page, nil
}

func (r *pagesRepository) Delete(ctx context.Context, userID, pageID uuid.UUID) error {
	query := `
		DELETE FROM content.pages
		WHERE id = $1
			AND project_id IN (SELECT id FROM content.projects WHERE user_id = $2)
	`

	result, err := r.db.Pool.Exec(ctx, query, pageID, userID)
	if err != nil {
		return utils.ErrDatabase("Failed to delete page", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrNotFound("Page")
	}

	return nil
}
