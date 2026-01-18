package services

import (
	"context"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
)

const (
	defaultCanvasWidth  = 1080
	defaultCanvasHeight = 1920
	defaultBackground   = "#FFFFFF"
)

type PagesService interface {
	CreatePage(ctx context.Context, userID uuid.UUID, page *models.Page) (*models.Page, error)
	ListPages(ctx context.Context, userID, projectID uuid.UUID, page, perPage int) ([]*models.Page, int, error)
	GetPage(ctx context.Context, userID, pageID uuid.UUID) (*models.Page, error)
	UpdatePage(ctx context.Context, userID uuid.UUID, update *models.PageUpdate) (*models.Page, error)
	DeletePage(ctx context.Context, userID, pageID uuid.UUID) error
}

type pagesService struct {
	pagesRepo repository.PagesRepository
}

func NewPagesService(pagesRepo repository.PagesRepository) PagesService {
	return &pagesService{pagesRepo: pagesRepo}
}

func (s *pagesService) CreatePage(ctx context.Context, userID uuid.UUID, page *models.Page) (*models.Page, error) {
	if page.PageNumber <= 0 {
		return nil, utils.ErrBadRequest("Page number must be greater than 0", nil)
	}

	if page.CanvasWidth <= 0 {
		page.CanvasWidth = defaultCanvasWidth
	}
	if page.CanvasHeight <= 0 {
		page.CanvasHeight = defaultCanvasHeight
	}
	if page.BackgroundColor == "" {
		page.BackgroundColor = defaultBackground
	}

	if page.ID == uuid.Nil {
		page.ID = uuid.New()
	}

	if err := s.pagesRepo.Create(ctx, userID, page); err != nil {
		return nil, err
	}

	return page, nil
}

func (s *pagesService) ListPages(ctx context.Context, userID, projectID uuid.UUID, pageNumber, perPage int) ([]*models.Page, int, error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (pageNumber - 1) * perPage

	return s.pagesRepo.ListByProject(ctx, userID, projectID, perPage, offset)
}

func (s *pagesService) GetPage(ctx context.Context, userID, pageID uuid.UUID) (*models.Page, error) {
	return s.pagesRepo.GetByID(ctx, userID, pageID)
}

func (s *pagesService) UpdatePage(ctx context.Context, userID uuid.UUID, update *models.PageUpdate) (*models.Page, error) {
	if update.PageNumber != nil && *update.PageNumber <= 0 {
		return nil, utils.ErrBadRequest("Page number must be greater than 0", nil)
	}

	updated, err := s.pagesRepo.Update(ctx, userID, update)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *pagesService) DeletePage(ctx context.Context, userID, pageID uuid.UUID) error {
	return s.pagesRepo.Delete(ctx, userID, pageID)
}
