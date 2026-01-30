package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
)

const (
	defaultSignedURLExpiry = 1 * time.Hour
)

type ItemsService interface {
	CreateItem(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier, fileHeader *multipart.FileHeader, format string, itemName string, itemCategory string, tags []string) (*models.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Item, int, error)
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error)
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error
	GetUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, map[string]string, error)
}

type itemsService struct {
	itemsRepo    repository.ItemsRepository
	usageService UsageService
	mlClient     MLClient
	storage      Storage
}

func NewItemsService(itemsRepo repository.ItemsRepository, usageService UsageService, mlClient MLClient, storage Storage) ItemsService {
	return &itemsService{
		itemsRepo:    itemsRepo,
		usageService: usageService,
		mlClient:     mlClient,
		storage:      storage,
	}
}

func (s *itemsService) CreateItem(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier, fileHeader *multipart.FileHeader, format string, itemName string, itemCategory string, tags []string) (*models.Item, error) {
	if fileHeader == nil {
		return nil, utils.ErrBadRequest("Image file is required", nil)
	}

	if err := utils.ValidateImageFile(fileHeader); err != nil {
		return nil, err
	}

	canProcess, _, err := s.usageService.CheckAndIncrementUsage(ctx, userID, tier)
	if err != nil {
		return nil, err
	}
	if !canProcess {
		return nil, utils.ErrRateLimitExceeded("Usage limit exceeded")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to open image file", err)
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to read image file", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	outputFormat := strings.ToLower(strings.TrimSpace(format))
	if outputFormat == "" {
		outputFormat = "png"
	}
	if outputFormat != "png" && outputFormat != "webp" {
		return nil, utils.ErrBadRequest("Unsupported output format", nil)
	}

	startedAt := time.Now().UTC()
	mlResponse, err := s.mlClient.RemoveBackground(ctx, base64Image, outputFormat)
	if err != nil {
		return nil, err
	}
	completedAt := time.Now().UTC()

	processedBytes, err := base64.StdEncoding.DecodeString(mlResponse.ProcessedImage)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to decode processed image", err)
	}

	originalKey, err := s.storage.Upload(ctx, bytes.NewReader(imageBytes), fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, utils.ErrStorage("Failed to upload original image", err)
	}

	processedContentType := formatToContentType(outputFormat)
	processedFilename := fmt.Sprintf("%s.%s", uuid.New().String(), outputFormat)
	processedKey, err := s.storage.Upload(ctx, bytes.NewReader(processedBytes), processedFilename, processedContentType)
	if err != nil {
		_ = s.storage.Delete(ctx, originalKey)
		return nil, utils.ErrStorage("Failed to upload processed image", err)
	}

	originalURL, err := s.storage.GetURL(ctx, originalKey, defaultSignedURLExpiry)
	if err != nil {
		_ = s.storage.Delete(ctx, originalKey)
		_ = s.storage.Delete(ctx, processedKey)
		return nil, utils.ErrStorage("Failed to generate original image URL", err)
	}

	processedURL, err := s.storage.GetURL(ctx, processedKey, defaultSignedURLExpiry)
	if err != nil {
		_ = s.storage.Delete(ctx, originalKey)
		_ = s.storage.Delete(ctx, processedKey)
		return nil, utils.ErrStorage("Failed to generate processed image URL", err)
	}

	itemID := uuid.New()
	processingStatus := "completed"
	modelVersion := mlResponse.Metadata.Model
	mimeType := fileHeader.Header.Get("Content-Type")

	var itemNamePtr *string
	var itemCategoryPtr *string
	if strings.TrimSpace(itemName) != "" {
		itemNamePtr = &itemName
	}
	if strings.TrimSpace(itemCategory) != "" {
		itemCategoryPtr = &itemCategory
	}

	originalSize := int64(len(imageBytes))
	item := &models.Item{
		ID:                   itemID,
		UserID:               userID,
		OriginalImageKey:     originalKey,
		OriginalImageURL:     originalURL,
		OriginalFileSize:     &originalSize,
		ProcessedImageKey:    &processedKey,
		ProcessedImageURL:    &processedURL,
		ProcessedFileSize:    int64Ptr(int64(len(processedBytes))),
		ProcessingStatus:     processingStatus,
		MLModelVersion:       stringPtr(modelVersion),
		ProcessingStartedAt:  &startedAt,
		ProcessingCompletedAt: &completedAt,
		MimeType:             stringPtr(mimeType),
		ItemName:             itemNamePtr,
		ItemCategory:         itemCategoryPtr,
		Tags:                 tags,
	}

	if err := s.itemsRepo.Create(ctx, item); err != nil {
		_ = s.storage.Delete(ctx, originalKey)
		_ = s.storage.Delete(ctx, processedKey)
		return nil, err
	}

	return item, nil
}

func (s *itemsService) ListItems(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Item, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	items, total, err := s.itemsRepo.ListByUser(ctx, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	for _, item := range items {
		signedOriginal, err := s.storage.GetURL(ctx, item.OriginalImageKey, defaultSignedURLExpiry)
		if err != nil {
			return nil, 0, utils.ErrStorage("Failed to generate original image URL", err)
		}
		item.OriginalImageURL = signedOriginal

		if item.ProcessedImageKey != nil {
			signedProcessed, err := s.storage.GetURL(ctx, *item.ProcessedImageKey, defaultSignedURLExpiry)
			if err != nil {
				return nil, 0, utils.ErrStorage("Failed to generate processed image URL", err)
			}
			item.ProcessedImageURL = &signedProcessed
		}
	}

	return items, total, nil
}

func (s *itemsService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	item, err := s.itemsRepo.GetByID(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}

	signedOriginal, err := s.storage.GetURL(ctx, item.OriginalImageKey, defaultSignedURLExpiry)
	if err != nil {
		return nil, utils.ErrStorage("Failed to generate original image URL", err)
	}
	item.OriginalImageURL = signedOriginal

	if item.ProcessedImageKey != nil {
		signedProcessed, err := s.storage.GetURL(ctx, *item.ProcessedImageKey, defaultSignedURLExpiry)
		if err != nil {
			return nil, utils.ErrStorage("Failed to generate processed image URL", err)
		}
		item.ProcessedImageURL = &signedProcessed
	}

	return item, nil
}

func (s *itemsService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	item, err := s.itemsRepo.SoftDelete(ctx, userID, itemID)
	if err != nil {
		return err
	}

	if err := s.storage.Delete(ctx, item.OriginalImageKey); err != nil {
		return utils.ErrStorage("Failed to delete original image", err)
	}

	if item.ProcessedImageKey != nil {
		if err := s.storage.Delete(ctx, *item.ProcessedImageKey); err != nil {
			return utils.ErrStorage("Failed to delete processed image", err)
		}
	}

	return nil
}

func (s *itemsService) GetUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, map[string]string, error) {
	stats, err := s.usageService.GetCurrentUsageStats(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	headers := s.usageService.GetRateLimitHeaders(stats)
	return stats, headers, nil
}

func formatToContentType(format string) string {
	switch strings.ToLower(format) {
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func int64Ptr(val int64) *int64 {
	return &val
}

func stringPtr(val string) *string {
	if strings.TrimSpace(val) == "" {
		return nil
	}
	return &val
}
