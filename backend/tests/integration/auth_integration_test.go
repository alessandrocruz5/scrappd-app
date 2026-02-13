//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/api"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
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

type noopItemsService struct{}

func (s *noopItemsService) CreateItem(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier, fileHeader *multipart.FileHeader, format string, itemName string, itemCategory string, tags []string) (*models.Item, error) {
	return nil, utils.ErrInternalServer("Items service not configured", nil)
}

func (s *noopItemsService) ListItems(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Item, int, error) {
	return nil, 0, utils.ErrInternalServer("Items service not configured", nil)
}

func (s *noopItemsService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	return nil, utils.ErrInternalServer("Items service not configured", nil)
}

func (s *noopItemsService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	return utils.ErrInternalServer("Items service not configured", nil)
}

func (s *noopItemsService) GetUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, map[string]string, error) {
	return nil, nil, utils.ErrInternalServer("Items service not configured", nil)
}

func (s *noopItemsService) ProcessItem(ctx context.Context, userID, itemID uuid.UUID, format string) error {
	return utils.ErrInternalServer("Items service not configured", nil)
}

func (s *noopItemsService) CancelProcessing(ctx context.Context, userID, itemID uuid.UUID) error {
	return utils.ErrInternalServer("Items service not configured", nil)
}

func TestAuthFlow_Integration(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	userRepo := repository.NewUserRepository(db)
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)
	authService := services.NewAuthService(userRepo, tokenManager, services.NoopEmailSender{}, "http://localhost:3000")
	mlClient := services.NewMLClient(&config.MLServiceConfig{
		BaseURL:    "http://localhost:8000",
		Timeout:    120 * time.Second,
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
	})

	router := api.SetupRouter(
		mlClient,
		authService,
		&noopItemsService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		tokenManager,
		"test-internal-secret",
		logger,
	)

	testEmail := "integration-test@example.com"
	testPassword := "TestPassword123!"

	// Cleanup before test
	cleanupTestUser(t, db, testEmail)
	defer cleanupTestUser(t, db, testEmail)

	// Test 1: Register new user
	t.Run("Register", func(t *testing.T) {
		reqBody := models.CreateUserRequest{
			Email:    testEmail,
			Username: "integrationtest",
			Password: testPassword,
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response utils.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	var accessToken, refreshToken string

	// Test 2: Login
	t.Run("Login", func(t *testing.T) {
		reqBody := models.LoginRequest{
			Email:    testEmail,
			Password: testPassword,
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response utils.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		// Extract tokens from response
		data := response.Data.(map[string]interface{})
		accessToken = data["access_token"].(string)
		refreshToken = data["refresh_token"].(string)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	// Test 3: Access protected endpoint
	t.Run("GetMe", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response utils.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, testEmail, data["email"])
	})

	// Test 4: Refresh token
	t.Run("RefreshToken", func(t *testing.T) {
		// Add 1 second delay to ensure different timestamp in JWT
		time.Sleep(1 * time.Second)

		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response utils.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		newAccessToken := data["access_token"].(string)
		newRefreshToken := data["refresh_token"].(string)

		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, accessToken, newAccessToken)
		assert.NotEqual(t, refreshToken, newRefreshToken)

		// Update tokens for next test
		accessToken = newAccessToken
		refreshToken = newRefreshToken
	})

	// Test 5: Logout
	t.Run("Logout", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response utils.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	// Test 6: Try to use revoked refresh token
	t.Run("UseRevokedToken", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test 7: Duplicate registration
	t.Run("DuplicateRegistration", func(t *testing.T) {
		reqBody := models.CreateUserRequest{
			Email:    testEmail,
			Username: "integrationtest2",
			Password: testPassword,
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	// Test 8: Wrong password
	t.Run("WrongPassword", func(t *testing.T) {
		reqBody := models.LoginRequest{
			Email:    testEmail,
			Password: "WrongPassword123!",
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
