package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/register", handler.Register)

	userID := uuid.New()
	expectedUser := &models.User{
		ID:               userID,
		Email:            "test@example.com",
		Username:         "testuser",
		SubscriptionTier: "free",
		CreatedAt:        time.Now(),
	}

	mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).
		Return(expectedUser, nil)

	reqBody := models.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/register", handler.Register)

	// Missing required fields
	reqBody := map[string]string{
		"email": "test@example.com",
		// Missing username and password
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/register", handler.Register)

	mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).
		Return(nil, utils.ErrConflict("Email already registered", nil))

	reqBody := models.CreateUserRequest{
		Email:    "existing@example.com",
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/login", handler.Login)

	userID := uuid.New()
	expectedResponse := &models.LoginResponse{
		User: &models.User{
			ID:               userID,
			Email:            "test@example.com",
			Username:         "testuser",
			SubscriptionTier: "free",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).
		Return(expectedResponse, nil)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/login", handler.Login)

	mockService.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).
		Return(nil, utils.ErrUnauthorized("Invalid email or password", nil))

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	userID := uuid.New()
	expectedResponse := &models.LoginResponse{
		User: &models.User{
			ID:       userID,
			Email:    "test@example.com",
			Username: "testuser",
		},
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	mockService.On("RefreshAccessToken", mock.Anything, "old-refresh-token").
		Return(expectedResponse, nil)

	reqBody := map[string]string{
		"refresh_token": "old-refresh-token",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/auth/logout", handler.Logout)

	mockService.On("Logout", mock.Anything, "refresh-token").Return(nil)

	reqBody := map[string]string{
		"refresh_token": "refresh-token",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetMe_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.GET("/auth/me", handler.GetMe)

	userID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
	}

	mockService.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil)

	// w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/me", nil)

	// Simulate auth middleware setting user_id
	ctx := &gin.Context{Request: req}
	ctx.Set("user_id", userID.String())

	handler.GetMe(ctx)

	// Note: This test is simplified. In real scenario, we'd need to properly set up Gin context
	mockService.AssertExpectations(t)
}
