package services

import (
	"context"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockUserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *MockUserRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	ctx := context.Background()
	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	// Mock: User doesn't exist
	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, assert.AnError)
	mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, assert.AnError)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	user, err := service.Register(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Username, user.Username)
	assert.NotEqual(t, req.Password, user.PasswordHash) // Password should be hashed
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := auth.HashPassword(password)

	existingUser := &models.User{
		ID:               uuid.New(),
		Email:            "test@example.com",
		Username:         "testuser",
		PasswordHash:     hashedPassword,
		SubscriptionTier: "free",
	}

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)
	mockRepo.On("CreateRefreshToken", ctx, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockRepo.On("UpdateLastLogin", ctx, existingUser.ID).Return(nil)

	response, err := service.Login(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, existingUser.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)
	service := NewAuthService(mockRepo, tokenManager)

	ctx := context.Background()
	hashedPassword, _ := auth.HashPassword("correctpassword")

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
	}

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	response, err := service.Login(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}
