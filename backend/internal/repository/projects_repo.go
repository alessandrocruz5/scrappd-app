package repository

import (
	"context"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectsRepository interface {
	Create(ctx context.Context, project *models.Project) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Project, int, error)
	GetByID(ctx context.Context, userID, projectID uuid.UUID) (*models.Project, error)
	Update(ctx context.Context, userID uuid.UUID, update *models.ProjectUpdate) (*models.Project, error)
	SoftDelete(ctx context.Context, userID, projectID uuid.UUID) error
}

type projectsRepository struct {
	db *database.DB
}

func NewProjectsRepository(db *database.DB) ProjectsRepository {
	return &projectsRepository{db: db}
}

func (r *projectsRepository) Create(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO content.projects (
			id,
			user_id,
			title,
			description,
			cover_image_url,
			visibility,
			is_template,
			template_price
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		project.ID,
		project.UserID,
		project.Title,
		project.Description,
		project.CoverImageURL,
		project.Visibility,
		project.IsTemplate,
		project.TemplatePrice,
	).Scan(&project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return utils.ErrDatabase("Failed to create project", err)
	}

	return nil
}

func (r *projectsRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Project, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	countQuery := `
		SELECT COUNT(*)
		FROM content.projects
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var total int
	if err := r.db.Pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, utils.ErrDatabase("Failed to count projects", err)
	}

	query := `
		SELECT
			id,
			user_id,
			title,
			description,
			cover_image_url,
			visibility,
			is_template,
			template_price,
			view_count,
			like_count,
			fork_count,
			created_at,
			updated_at,
			published_at
		FROM content.projects
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrDatabase("Failed to list projects", err)
	}
	defer rows.Close()

	projects := []*models.Project{}
	for rows.Next() {
		project := &models.Project{}
		if err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Title,
			&project.Description,
			&project.CoverImageURL,
			&project.Visibility,
			&project.IsTemplate,
			&project.TemplatePrice,
			&project.ViewCount,
			&project.LikeCount,
			&project.ForkCount,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.PublishedAt,
		); err != nil {
			return nil, 0, utils.ErrDatabase("Failed to scan projects", err)
		}
		projects = append(projects, project)
	}

	if rows.Err() != nil {
		return nil, 0, utils.ErrDatabase("Failed to iterate projects", rows.Err())
	}

	return projects, total, nil
}

func (r *projectsRepository) GetByID(ctx context.Context, userID, projectID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT
			id,
			user_id,
			title,
			description,
			cover_image_url,
			visibility,
			is_template,
			template_price,
			view_count,
			like_count,
			fork_count,
			created_at,
			updated_at,
			published_at
		FROM content.projects
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	project := &models.Project{}
	err := r.db.Pool.QueryRow(ctx, query, projectID, userID).Scan(
		&project.ID,
		&project.UserID,
		&project.Title,
		&project.Description,
		&project.CoverImageURL,
		&project.Visibility,
		&project.IsTemplate,
		&project.TemplatePrice,
		&project.ViewCount,
		&project.LikeCount,
		&project.ForkCount,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.PublishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Project")
		}
		return nil, utils.ErrDatabase("Failed to get project", err)
	}

	return project, nil
}

func (r *projectsRepository) Update(ctx context.Context, userID uuid.UUID, update *models.ProjectUpdate) (*models.Project, error) {
	query := `
		UPDATE content.projects
		SET
			title = COALESCE($2, title),
			description = COALESCE($3, description),
			cover_image_url = COALESCE($4, cover_image_url),
			visibility = COALESCE($5, visibility),
			is_template = COALESCE($6, is_template),
			template_price = COALESCE($7, template_price),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $8 AND deleted_at IS NULL
		RETURNING
			id,
			user_id,
			title,
			description,
			cover_image_url,
			visibility,
			is_template,
			template_price,
			view_count,
			like_count,
			fork_count,
			created_at,
			updated_at,
			published_at
	`

	project := &models.Project{}
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		update.ID,
		update.Title,
		update.Description,
		update.CoverImageURL,
		update.Visibility,
		update.IsTemplate,
		update.TemplatePrice,
		userID,
	).Scan(
		&project.ID,
		&project.UserID,
		&project.Title,
		&project.Description,
		&project.CoverImageURL,
		&project.Visibility,
		&project.IsTemplate,
		&project.TemplatePrice,
		&project.ViewCount,
		&project.LikeCount,
		&project.ForkCount,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.PublishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNotFound("Project")
		}
		return nil, utils.ErrDatabase("Failed to update project", err)
	}

	return project, nil
}

func (r *projectsRepository) SoftDelete(ctx context.Context, userID, projectID uuid.UUID) error {
	query := `
		UPDATE content.projects
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Pool.Exec(ctx, query, projectID, userID)
	if err != nil {
		return utils.ErrDatabase("Failed to delete project", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrNotFound("Project")
	}

	return nil
}
