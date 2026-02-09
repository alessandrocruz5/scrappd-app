//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/api"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMLClient struct{}

func (s *stubMLClient) RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error) {
	processed := base64.StdEncoding.EncodeToString([]byte("processed"))
	return &models.RemoveBackgroundResponse{
		ProcessedImage: processed,
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 0.1,
			Model:          "birefnet-general",
		},
	}, nil
}

func (s *stubMLClient) HealthCheck(ctx context.Context) (*models.HealthCheckResponse, error) {
	return &models.HealthCheckResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Model:   "birefnet-general",
		Time:    time.Now(),
	}, nil
}

type inMemoryStorage struct {
	keys map[string][]byte
}

func newInMemoryStorage() *inMemoryStorage {
	return &inMemoryStorage{keys: make(map[string][]byte)}
}

func (s *inMemoryStorage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	key := "uploads/" + uuid.New().String()
	s.keys[key] = data
	return key, nil
}

func (s *inMemoryStorage) UploadWithKey(ctx context.Context, file io.Reader, key string, contentType string) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	s.keys[key] = data
	return nil
}

func (s *inMemoryStorage) Download(ctx context.Context, key string) ([]byte, error) {
	return s.keys[key], nil
}

func (s *inMemoryStorage) Delete(ctx context.Context, key string) error {
	delete(s.keys, key)
	return nil
}

func (s *inMemoryStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "signed://" + key, nil
}

func (s *inMemoryStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := s.keys[key]
	return ok, nil
}

func (s *inMemoryStorage) List(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}

func (s *inMemoryStorage) HealthCheck(ctx context.Context) error {
	return nil
}

type stubRedis struct{}

func (s *stubRedis) Ping(ctx context.Context) error {
	return nil
}

func (s *stubRedis) Close() error {
	return nil
}

func TestItemsFlow_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	userRepo := repository.NewUserRepository(db)
	usageRepo := repository.NewUsageRepository(db.Pool)
	itemsRepo := repository.NewItemsRepository(db)
	pagesRepo := repository.NewPagesRepository(db)
	projectsRepo := repository.NewProjectsRepository(db)
	pageItemsRepo := repository.NewPageItemsRepository(db)

	user := buildTestUser("items-flow@test.com", "itemsflow", models.TierFree)
	err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)
	defer cleanupTestUser(t, db, user.Email)

	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)
	accessToken, err := tokenManager.GenerateAccessToken(user.ID, user.Email, user.Username, string(user.SubscriptionTier))
	require.NoError(t, err)

	usageService := services.NewUsageService(usageRepo)
	mlClient := &stubMLClient{}
	storage := newInMemoryStorage()
	itemsService := services.NewItemsService(itemsRepo, usageService, mlClient, storage, false)
	authService := services.NewAuthService(userRepo, tokenManager)
	pagesService := services.NewPagesService(pagesRepo)
	projectsService := services.NewProjectsService(projectsRepo)
	pageItemsService := services.NewPageItemsService(pageItemsRepo)
	pageRenderService := services.NewPageRenderService(pagesRepo, pageItemsRepo, itemsRepo, storage)

	router := api.SetupRouter(
		mlClient,
		authService,
		itemsService,
		projectsService,
		pagesService,
		pageItemsService,
		pageRenderService,
		usageService,
		db,
		&stubRedis{},
		storage,
		tokenManager,
		logger,
	)

	t.Run("create_item", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("image", "test.png")
		require.NoError(t, err)
		_, err = part.Write([]byte("image-bytes"))
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		req := httptest.NewRequest(http.MethodPost, "/api/v1/items", body)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response utils.Response
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("list_items_with_rate_limit_headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	})
}

func buildTestUser(email, username string, tier models.SubscriptionTier) *models.User {
	return &models.User{
		ID:                     uuid.New(),
		Email:                  email,
		Username:               username,
		PasswordHash:           "$2a$10$validhashhere",
		SubscriptionTier:       tier,
		SubscriptionStatus:     "active",
		MonthlyBgRemovalsUsed:  0,
		MonthlyBgRemovalsLimit: 5,
		MonthlyStorageUsedMB:   0,
		MonthlyStorageLimitMB:  100,
		FollowerCount:          0,
		FollowingCount:         0,
		IsVerified:             false,
	}
}
