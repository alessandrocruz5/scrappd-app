package services

import (
	"context"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
)

type PageItemsService interface {
	CreatePageItem(ctx context.Context, userID uuid.UUID, item *models.PageItem) (*models.PageItem, error)
	ListPageItems(ctx context.Context, userID, pageID uuid.UUID) ([]*models.PageItem, error)
	UpdatePageItem(ctx context.Context, userID uuid.UUID, update *models.PageItemUpdate) (*models.PageItem, error)
	DeletePageItem(ctx context.Context, userID, pageID, pageItemID uuid.UUID) error
}

type pageItemsService struct {
	pageItemsRepo repository.PageItemsRepository
}

func NewPageItemsService(pageItemsRepo repository.PageItemsRepository) PageItemsService {
	return &pageItemsService{pageItemsRepo: pageItemsRepo}
}

func (s *pageItemsService) CreatePageItem(ctx context.Context, userID uuid.UUID, item *models.PageItem) (*models.PageItem, error) {
	if item.PageID == uuid.Nil || item.ItemID == uuid.Nil {
		return nil, utils.ErrBadRequest("Page ID and item ID are required", nil)
	}
	if item.Width <= 0 || item.Height <= 0 {
		return nil, utils.ErrBadRequest("Width and height must be greater than 0", nil)
	}
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.Opacity == 0 {
		item.Opacity = 1.0
	}

	if err := s.pageItemsRepo.Create(ctx, userID, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *pageItemsService) ListPageItems(ctx context.Context, userID, pageID uuid.UUID) ([]*models.PageItem, error) {
	return s.pageItemsRepo.ListByPage(ctx, userID, pageID)
}

func (s *pageItemsService) UpdatePageItem(ctx context.Context, userID uuid.UUID, update *models.PageItemUpdate) (*models.PageItem, error) {
	if update.PageID == uuid.Nil || update.ID == uuid.Nil {
		return nil, utils.ErrBadRequest("Page ID and page item ID are required", nil)
	}

	if update.Width != nil && *update.Width <= 0 {
		return nil, utils.ErrBadRequest("Width must be greater than 0", nil)
	}
	if update.Height != nil && *update.Height <= 0 {
		return nil, utils.ErrBadRequest("Height must be greater than 0", nil)
	}

	return s.pageItemsRepo.Update(ctx, userID, update)
}

func (s *pageItemsService) DeletePageItem(ctx context.Context, userID, pageID, pageItemID uuid.UUID) error {
	return s.pageItemsRepo.Delete(ctx, userID, pageID, pageItemID)
}
