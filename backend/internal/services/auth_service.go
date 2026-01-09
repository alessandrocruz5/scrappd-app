package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type authService struct {
	userRepo     repository.UserRepository
	tokenManager *auth.TokenManager
}

func NewAuthService(userRepo repository.UserRepository, tokenManager *auth.TokenManager) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *authService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, utils.ErrConflict("Email already registered", nil)
	}

	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, utils.ErrConflict("Username already taken", nil)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to hash password", err)
	}

	// Create user
	user := &models.User{
		ID:                     uuid.New(),
		Email:                  req.Email,
		Username:               req.Username,
		DisplayName:            req.DisplayName,
		PasswordHash:           hashedPassword,
		SubscriptionTier:       "free",
		SubscriptionStatus:     "active",
		MonthlyBgRemovalsLimit: 5,   // Free tier: 5 removals/month
		MonthlyStorageLimitMB:  100, // Free tier: 100MB storage
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, utils.ErrUnauthorized("Invalid email or password", nil)
	}

	// Check password
	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, utils.ErrUnauthorized("Invalid email or password", nil)
	}

	// Generate access token
	accessToken, err := s.tokenManager.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Username,
		user.SubscriptionTier,
	)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to generate access token", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to generate refresh token", err)
	}

	// Store refresh token hash in database
	tokenHash := hashToken(refreshToken)
	refreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.userRepo.CreateRefreshToken(ctx, refreshTokenModel); err != nil {
		return nil, err
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return &models.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, utils.ErrUnauthorized("Invalid refresh token", err)
	}

	// Check if token exists in database and is not revoked
	tokenHash := hashToken(refreshToken)
	storedToken, err := s.userRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, utils.ErrUnauthorized("Invalid refresh token", err)
	}

	// Check if token is expired
	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, utils.ErrUnauthorized("Refresh token expired", nil)
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, utils.ErrUnauthorized("Invalid token claims", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := s.tokenManager.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Username,
		user.SubscriptionTier,
	)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to generate access token", err)
	}

	// Optionally: Generate new refresh token (rotate)
	newRefreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to generate refresh token", err)
	}

	// Revoke old refresh token
	if err := s.userRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		fmt.Printf("Failed to revoke old refresh token: %v\n", err)
	}

	// Store new refresh token
	newTokenHash := hashToken(newRefreshToken)
	newRefreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.userRepo.CreateRefreshToken(ctx, newRefreshTokenModel); err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.userRepo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *authService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// hashToken creates a SHA-256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
