package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMLClient is a mock implementation of the MLClient interface
type MockMLClient struct {
	mock.Mock
}

func (m *MockMLClient) RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error) {
	args := m.Called(ctx, imageData, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RemoveBackgroundResponse), args.Error(1)
}

func (m *MockMLClient) HealthCheck(ctx context.Context) (*models.HealthCheckResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HealthCheckResponse), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

type MockDBHealth struct {
	mock.Mock
}

func (m *MockDBHealth) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockRedisHealth struct {
	mock.Mock
}

func (m *MockRedisHealth) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockStorageHealth struct {
	mock.Mock
}

func (m *MockStorageHealth) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestBasicHealth(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewHealthHandler(mockMLClient, nil, nil, nil)

	router := setupTestRouter()
	router.GET("/health", handler.BasicHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Check the health response structure
	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusHealthy, data["status"])
	assert.NotNil(t, data["timestamp"])
	assert.Equal(t, "1.0.0", data["version"])

	mockMLClient.AssertNotCalled(t, "HealthCheck")
}

func TestDeepHealth_AllServicesHealthy(t *testing.T) {
	mockMLClient := new(MockMLClient)
	mockDB := new(MockDBHealth)
	mockRedis := new(MockRedisHealth)
	mockStorage := new(MockStorageHealth)

	mockDB.On("Health", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("HealthCheck", mock.Anything).Return(nil)
	mockMLClient.On("HealthCheck", mock.Anything).Return(&models.HealthCheckResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Model:   "BiRefNet",
		Time:    time.Now(),
	}, nil)

	handler := NewHealthHandler(mockMLClient, mockDB, mockRedis, mockStorage)

	router := setupTestRouter()
	router.GET("/health/deep", handler.DeepHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/deep", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusHealthy, data["status"])

	services := data["services"].(map[string]interface{})
	mlService := services["ml_service"].(map[string]interface{})
	assert.Equal(t, StatusHealthy, mlService["status"])
	assert.Equal(t, "ML service is operational", mlService["message"])

	mockMLClient.AssertExpectations(t)
}

func TestDeepHealth_MLServiceUnhealthy(t *testing.T) {
	mockMLClient := new(MockMLClient)
	mockDB := new(MockDBHealth)
	mockRedis := new(MockRedisHealth)
	mockStorage := new(MockStorageHealth)

	mockDB.On("Health", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("HealthCheck", mock.Anything).Return(nil)
	mockMLClient.On("HealthCheck", mock.Anything).Return(
		nil,
		utils.ErrServiceUnavailable("ML", errors.New("connection refused")),
	)

	handler := NewHealthHandler(mockMLClient, mockDB, mockRedis, mockStorage)

	router := setupTestRouter()
	router.GET("/health/deep", handler.DeepHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/deep", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success) // Still returns success structure, just with degraded status
	assert.NotNil(t, response.Data)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusDegraded, data["status"])

	services := data["services"].(map[string]interface{})
	mlService := services["ml_service"].(map[string]interface{})
	assert.Equal(t, StatusUnhealthy, mlService["status"])
	assert.Equal(t, "ML service is unreachable", mlService["message"])

	mockMLClient.AssertExpectations(t)
}

func TestDeepHealth_MLServiceReportsUnhealthy(t *testing.T) {
	mockMLClient := new(MockMLClient)
	mockDB := new(MockDBHealth)
	mockRedis := new(MockRedisHealth)
	mockStorage := new(MockStorageHealth)

	mockDB.On("Health", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("HealthCheck", mock.Anything).Return(nil)
	mockMLClient.On("HealthCheck", mock.Anything).Return(&models.HealthCheckResponse{
		Status:  "unhealthy",
		Version: "1.0.0",
		Model:   "BiRefNet",
		Time:    time.Now(),
	}, nil)

	handler := NewHealthHandler(mockMLClient, mockDB, mockRedis, mockStorage)

	router := setupTestRouter()
	router.GET("/health/deep", handler.DeepHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/deep", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusDegraded, data["status"])

	services := data["services"].(map[string]interface{})
	mlService := services["ml_service"].(map[string]interface{})
	assert.Equal(t, StatusUnhealthy, mlService["status"])
	assert.Contains(t, mlService["message"], "unhealthy status")

	mockMLClient.AssertExpectations(t)
}

func TestReadinessProbe_Ready(t *testing.T) {
	mockMLClient := new(MockMLClient)
	mockDB := new(MockDBHealth)
	mockRedis := new(MockRedisHealth)
	mockStorage := new(MockStorageHealth)

	mockDB.On("Health", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("HealthCheck", mock.Anything).Return(nil)
	mockMLClient.On("HealthCheck", mock.Anything).Return(&models.HealthCheckResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Model:   "BiRefNet",
		Time:    time.Now(),
	}, nil)

	handler := NewHealthHandler(mockMLClient, mockDB, mockRedis, mockStorage)

	router := setupTestRouter()
	router.GET("/health/ready", handler.ReadinessProbe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusHealthy, data["status"])

	mockMLClient.AssertExpectations(t)
}

func TestReadinessProbe_NotReady(t *testing.T) {
	mockMLClient := new(MockMLClient)
	mockDB := new(MockDBHealth)
	mockRedis := new(MockRedisHealth)
	mockStorage := new(MockStorageHealth)

	mockDB.On("Health", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("HealthCheck", mock.Anything).Return(nil)
	mockMLClient.On("HealthCheck", mock.Anything).Return(
		nil,
		utils.ErrServiceUnavailable("ML", errors.New("connection timeout")),
	)

	handler := NewHealthHandler(mockMLClient, mockDB, mockRedis, mockStorage)

	router := setupTestRouter()
	router.GET("/health/ready", handler.ReadinessProbe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, utils.ErrCodeServiceUnavailable, response.Error.Code)

	mockMLClient.AssertExpectations(t)
}

func TestLivenessProbe(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewHealthHandler(mockMLClient, nil, nil, nil)

	router := setupTestRouter()
	router.GET("/health/live", handler.LivenessProbe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/live", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusHealthy, data["status"])

	// Liveness probe should not call external services
	mockMLClient.AssertNotCalled(t, "HealthCheck")
}

func TestHealthHandler_ContextTimeout(t *testing.T) {
	mockMLClient := new(MockMLClient)
	mockDB := new(MockDBHealth)
	mockRedis := new(MockRedisHealth)
	mockStorage := new(MockStorageHealth)

	mockDB.On("Health", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("HealthCheck", mock.Anything).Return(nil)
	mockMLClient.On("HealthCheck", mock.Anything).Run(func(args mock.Arguments) {
		// Simulate slow response
		ctx := args.Get(0).(context.Context)
		select {
		case <-ctx.Done():
			// Context cancelled/timed out
		case <-time.After(20 * time.Second):
			// This should not be reached due to context timeout
		}
	}).Return(nil, context.DeadlineExceeded)

	handler := NewHealthHandler(mockMLClient, mockDB, mockRedis, mockStorage)

	router := setupTestRouter()
	router.GET("/health/deep", handler.DeepHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/deep", nil)
	router.ServeHTTP(w, req)

	// Should still return a response even if ML service times out
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, StatusDegraded, data["status"])
}
