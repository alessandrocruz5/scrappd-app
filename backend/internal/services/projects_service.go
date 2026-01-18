package services

import (
	"context"
	"strings"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
)

const defaultProjectVisibility = "private"

type ProjectsService interface {
	CreateProject(ctx context.Context, userID uuid.UUID, project *models.Project) (*models.Project, error)
	ListProjects(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Project, int, error)
	GetProject(ctx context.Context, userID, projectID uuid.UUID) (*models.Project, error)
	UpdateProject(ctx context.Context, userID uuid.UUID, update *models.ProjectUpdate) (*models.Project, error)
	DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error
}

type projectsService struct {
	projectsRepo repository.ProjectsRepository
}

func NewProjectsService(projectsRepo repository.ProjectsRepository) ProjectsService {
	return &projectsService{projectsRepo: projectsRepo}
}

func (s *projectsService) CreateProject(ctx context.Context, userID uuid.UUID, project *models.Project) (*models.Project, error) {
	if strings.TrimSpace(project.Title) == "" {
		return nil, utils.ErrBadRequest("Project title is required", nil)
	}

	if project.Visibility == "" {
		project.Visibility = defaultProjectVisibility
	}

	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	project.UserID = userID

	if err := s.projectsRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectsService) ListProjects(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Project, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	return s.projectsRepo.ListByUser(ctx, userID, perPage, offset)
}

func (s *projectsService) GetProject(ctx context.Context, userID, projectID uuid.UUID) (*models.Project, error) {
	return s.projectsRepo.GetByID(ctx, userID, projectID)
}

func (s *projectsService) UpdateProject(ctx context.Context, userID uuid.UUID, update *models.ProjectUpdate) (*models.Project, error) {
	if update.Title != nil && strings.TrimSpace(*update.Title) == "" {
		return nil, utils.ErrBadRequest("Project title cannot be empty", nil)
	}

	return s.projectsRepo.Update(ctx, userID, update)
}

func (s *projectsService) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	return s.projectsRepo.SoftDelete(ctx, userID, projectID)
}
