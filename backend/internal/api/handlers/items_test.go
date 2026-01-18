package handlers

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockItemsService struct {
	mock.Mock
}

func (m *mockItemsService) CreateItem(ctx context.Context, userID uuid.UUID, tier models.SubscriptionTier, fileHeader *multipart.FileHeader, format string, itemName string, itemCategory string, tags []string) (*models.Item, error) {
	args := m.Called(ctx, userID, tier, fileHeader, format, itemName, itemCategory, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *mockItemsService) ListItems(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Item, int, error) {
	args := m.Called(ctx, userID, page, perPage)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Item), args.Int(1), args.Error(2)
}

func (m *mockItemsService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *mockItemsService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	args := m.Called(ctx, userID, itemID)
	return args.Error(0)
}

func (m *mockItemsService) GetUsageStats(ctx context.Context, userID uuid.UUID) (*models.UsageStats, map[string]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*models.UsageStats), args.Get(1).(map[string]string), args.Error(2)
}

func TestItemsHandler_CreateItem_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	itemsService := new(mockItemsService)
	handler := NewItemsHandler(itemsService)
	router.Use(testAuthMiddleware(uuid.New(), models.TierFree))
	router.POST("/api/v1/items", handler.CreateItem)

	req, _ := http.NewRequest("POST", "/api/v1/items", nil)
	req.Header.Set("Content-Type", "multipart/form-data")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
}

func TestItemsHandler_ListItems_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	itemsService := new(mockItemsService)
	handler := NewItemsHandler(itemsService)
	userID := uuid.New()
	router.Use(testAuthMiddleware(userID, models.TierFree))
	router.GET("/api/v1/items", handler.ListItems)

	itemsService.On("ListItems", mock.Anything, userID, 1, 20).Return([]*models.Item{
		{ID: uuid.New(), UserID: userID, OriginalImageKey: "k", OriginalImageURL: "u"},
	}, 1, nil)

	req, _ := http.NewRequest("GET", "/api/v1/items", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestItemsHandler_GetUsage_SetsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	itemsService := new(mockItemsService)
	handler := NewItemsHandler(itemsService)
	userID := uuid.New()
	router.Use(testAuthMiddleware(userID, models.TierFree))
	router.GET("/api/v1/items/usage", handler.GetUsage)

	headers := map[string]string{
		"X-RateLimit-Limit":     "5",
		"X-RateLimit-Remaining": "4",
	}

	itemsService.On("GetUsageStats", mock.Anything, userID).Return(&models.UsageStats{}, headers, nil)

	req, _ := http.NewRequest("GET", "/api/v1/items/usage", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
}

func testAuthMiddleware(userID uuid.UUID, tier models.SubscriptionTier) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID.String())
		c.Set("subscription_tier", string(tier))
		c.Next()
	}
}
