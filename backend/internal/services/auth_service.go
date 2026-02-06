package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	RequestPasswordReset(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
	ResendVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
}

type authService struct {
	userRepo     repository.UserRepository
	tokenManager *auth.TokenManager
	emailSender  EmailSender
	appBaseURL   string
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenManager *auth.TokenManager,
	emailSender EmailSender,
	appBaseURL string,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		emailSender:  emailSender,
		appBaseURL:   appBaseURL,
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
		SubscriptionTier:       models.TierFree,
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
		string(user.SubscriptionTier),
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
		string(user.SubscriptionTier),
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

func (s *authService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) && appErr.Code == utils.ErrCodeNotFound {
			// Return success for unknown emails to avoid account enumeration.
			return "", nil
		}
		return "", err
	}

	rawToken, err := generateResetToken()
	if err != nil {
		return "", utils.ErrInternalServer("Failed to generate reset token", err)
	}

	reset := &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.userRepo.CreatePasswordReset(ctx, reset); err != nil {
		return "", err
	}

	if err := s.sendPasswordResetEmail(ctx, user.Email, rawToken, reset.ExpiresAt); err != nil {
		return "", utils.ErrServiceUnavailable("email", err)
	}

	return rawToken, nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	reset, err := s.userRepo.GetActivePasswordResetByTokenHash(ctx, hashToken(token))
	if err != nil {
		return err
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return utils.ErrInternalServer("Failed to hash password", err)
	}

	if err := s.userRepo.UpdatePasswordHash(ctx, reset.UserID, hashedPassword); err != nil {
		return err
	}
	if err := s.userRepo.MarkPasswordResetUsed(ctx, reset.ID); err != nil {
		return err
	}

	// Invalidate all active sessions after password reset.
	if err := s.userRepo.RevokeAllUserTokens(ctx, reset.UserID); err != nil {
		return err
	}

	return nil
}

func (s *authService) ResendVerification(ctx context.Context, email string) error {
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) && appErr.Code == utils.ErrCodeNotFound {
			// Return success for unknown emails to avoid account enumeration.
			return nil
		}
		return err
	}

	if err := s.sendVerificationEmail(ctx, email); err != nil {
		return utils.ErrServiceUnavailable("email", err)
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	claims, err := s.tokenManager.ValidateVerifyToken(token)
	if err != nil {
		return utils.ErrUnauthorized("Invalid or expired verification token", err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return utils.ErrBadRequest("Invalid verification token", err)
	}

	return s.userRepo.MarkEmailVerified(ctx, userID)
}

// hashToken creates a SHA-256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *authService) sendPasswordResetEmail(
	ctx context.Context,
	email string,
	token string,
	expiresAt time.Time,
) error {
	if s.emailSender == nil {
		return nil
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.appBaseURL, token)
	body := fmt.Sprintf(
		"Hi,\n\nWe received a request to reset your Scrapp'd password.\n\n"+
			"Reset link: %s\n\n"+
			"Reset token: %s\n\n"+
			"This token expires at %s.\n\nIf you didn't request this, you can ignore this email.",
		resetLink,
		token,
		expiresAt.Format(time.RFC1123),
	)

	return s.emailSender.Send(ctx, email, "Reset your Scrapp'd password", body)
}

func (s *authService) sendVerificationEmail(ctx context.Context, email string) error {
	if s.emailSender == nil {
		return nil
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) && appErr.Code == utils.ErrCodeNotFound {
			return nil
		}
		return err
	}

	token, err := s.tokenManager.GenerateVerifyToken(user.ID)
	if err != nil {
		return err
	}

	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", s.appBaseURL, token)
	body := fmt.Sprintf(
		"Hi,\n\nPlease verify your Scrapp'd email.\n\n"+
			"Verification link: %s\n\n"+
			"Verification token: %s\n\n"+
			"If you didn't request this, you can ignore this email.",
		verifyLink,
		token,
	)

	return s.emailSender.Send(ctx, email, "Verify your Scrapp'd email", body)
}
