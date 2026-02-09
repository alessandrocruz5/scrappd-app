package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockItemsRepo struct {
	mock.Mock
}

func (m *mockItemsRepo) Create(ctx context.Context, item *models.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockItemsRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Item, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Item), args.Int(1), args.Error(2)
}

func (m *mockItemsRepo) GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *mockItemsRepo) SoftDelete(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

type mockUsageService struct {
	mock.Mock
}

func (m *mockUsageService) CheckAndIncrementUsage(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) (bool, *models.UsageStats, error) {
	args := m.Called(ctx, userID, tier)
	return args.Bool(0), args.Get(1).(*models.UsageStats), args.Error(2)
}

func (m *mockUsageService) GetCurrentUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UsageStats), args.Error(1)
}

func (m *mockUsageService) GetUsageHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.UsageTracking, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]models.UsageTracking), args.Error(1)
}

func (m *mockUsageService) UpdateSubscriptionTier(ctx context.Context, userID uuid.UUID, newTier models.SubscriptionTier) error {
	args := m.Called(ctx, userID, newTier)
	return args.Error(0)
}

func (m *mockUsageService) ResetExpiredPeriods(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockUsageService) InitializeUsageForNewUser(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier) error {
	args := m.Called(ctx, userID, tier)
	return args.Error(0)
}

func (m *mockUsageService) GetRateLimitHeaders(stats *models.UsageStats) map[string]string {
	args := m.Called(stats)
	return args.Get(0).(map[string]string)
}

type mockMLClient struct {
	mock.Mock
}

func (m *mockMLClient) RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error) {
	args := m.Called(ctx, imageData, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RemoveBackgroundResponse), args.Error(1)
}

func (m *mockMLClient) HealthCheck(ctx context.Context) (*models.HealthCheckResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HealthCheckResponse), args.Error(1)
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	args := m.Called(ctx, filename, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) UploadWithKey(ctx context.Context, file io.Reader, key string, contentType string) error {
	args := m.Called(ctx, key, contentType)
	return args.Error(0)
}

func (m *mockStorage) Download(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *mockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockStorage) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestItemsService_CreateItem_Success(t *testing.T) {
	repo := new(mockItemsRepo)
	usage := new(mockUsageService)
	ml := new(mockMLClient)
	storage := new(mockStorage)
	service := NewItemsService(repo, usage, ml, storage, false)

	ctx := context.Background()
	userID := uuid.New()
	tier := models.TierFree

	fileHeader := buildTestFileHeader(t, "image.png", []byte("image-bytes"), "image/png")
	base64Image := base64.StdEncoding.EncodeToString([]byte("image-bytes"))
	processed := base64.StdEncoding.EncodeToString([]byte("processed"))

	usage.On("CheckAndIncrementUsage", ctx, userID, tier).Return(true, &models.UsageStats{}, nil)
	ml.On("RemoveBackground", ctx, base64Image, "png").Return(&models.RemoveBackgroundResponse{
		ProcessedImage: processed,
		Metadata: models.BackgroundRemovalMeta{
			Model: "birefnet-general",
		},
	}, nil)

	storage.On("Upload", ctx, "image.png", "image/png").Return("uploads/original.png", nil).Once()
	storage.On("Upload", ctx, mock.Anything, "image/png").Return("uploads/processed.png", nil).Once()
	storage.On("GetURL", ctx, "uploads/original.png", mock.Anything).Return("signed-original", nil).Once()
	storage.On("GetURL", ctx, "uploads/processed.png", mock.Anything).Return("signed-processed", nil).Once()

	repo.On("Create", ctx, mock.AnythingOfType("*models.Item")).Return(nil)

	item, err := service.CreateItem(ctx, userID, tier, fileHeader, "png", "", "", nil)
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "uploads/original.png", item.OriginalImageKey)
	assert.Equal(t, "signed-original", item.OriginalImageURL)
	assert.Equal(t, "signed-processed", *item.ProcessedImageURL)

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
	ml.AssertExpectations(t)
	usage.AssertExpectations(t)
}

func TestItemsService_CreateItem_UsageExceeded(t *testing.T) {
	repo := new(mockItemsRepo)
	usage := new(mockUsageService)
	ml := new(mockMLClient)
	storage := new(mockStorage)
	service := NewItemsService(repo, usage, ml, storage, false)

	ctx := context.Background()
	userID := uuid.New()

	fileHeader := buildTestFileHeader(t, "image.png", []byte("image-bytes"), "image/png")

	usage.On("CheckAndIncrementUsage", ctx, userID, models.TierFree).Return(false, &models.UsageStats{}, nil)

	item, err := service.CreateItem(ctx, userID, models.TierFree, fileHeader, "png", "", "", nil)
	require.Error(t, err)
	assert.Nil(t, item)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeRateLimitExceeded, appErr.Code)
}

func TestItemsService_ListItems_SignsURLs(t *testing.T) {
	repo := new(mockItemsRepo)
	usage := new(mockUsageService)
	ml := new(mockMLClient)
	storage := new(mockStorage)
	service := NewItemsService(repo, usage, ml, storage, false)

	ctx := context.Background()
	userID := uuid.New()

	processedKey := "processed-key"
	items := []*models.Item{
		{
			ID:                uuid.New(),
			UserID:            userID,
			OriginalImageKey:  "original-key",
			OriginalImageURL:  "stale-original",
			ProcessedImageKey: &processedKey,
			ProcessedImageURL: strPtr("stale-processed"),
		},
	}

	repo.On("ListByUser", ctx, userID, 20, 0).Return(items, 1, nil)
	storage.On("GetURL", ctx, "original-key", mock.Anything).Return("signed-original", nil).Once()
	storage.On("GetURL", ctx, processedKey, mock.Anything).Return("signed-processed", nil).Once()

	result, total, err := service.ListItems(ctx, userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "signed-original", result[0].OriginalImageURL)
	assert.Equal(t, "signed-processed", *result[0].ProcessedImageURL)
}

func TestItemsService_DeleteItem_DeletesStorage(t *testing.T) {
	repo := new(mockItemsRepo)
	usage := new(mockUsageService)
	ml := new(mockMLClient)
	storage := new(mockStorage)
	service := NewItemsService(repo, usage, ml, storage, false)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()
	processedKey := "processed-key"

	repo.On("SoftDelete", ctx, userID, itemID).Return(&models.Item{
		ID:                itemID,
		UserID:            userID,
		OriginalImageKey:  "original-key",
		ProcessedImageKey: &processedKey,
	}, nil)
	storage.On("Delete", ctx, "original-key").Return(nil)
	storage.On("Delete", ctx, processedKey).Return(nil)

	err := service.DeleteItem(ctx, userID, itemID)
	require.NoError(t, err)
}

func strPtr(val string) *string {
	return &val
}

func buildTestFileHeader(t *testing.T, filename string, data []byte, contentType string) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filename)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(int64(len(data))))

	files := req.MultipartForm.File["image"]
	require.NotEmpty(t, files)

	fileHeader := files[0]
	fileHeader.Header.Set("Content-Type", contentType)
	return fileHeader
}
