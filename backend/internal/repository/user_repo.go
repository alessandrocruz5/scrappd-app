package repository

import (
	"context"
	"errors"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Refresh token methods
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

type userRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO auth.users (
            id, email, username, display_name, password_hash,
            subscription_tier, subscription_status,
            monthly_bg_removals_limit, monthly_storage_limit_mb
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING created_at, updated_at
    `

	err := r.db.Pool.QueryRow(
		ctx, query,
		user.ID, user.Email, user.Username, user.DisplayName, user.PasswordHash,
		user.SubscriptionTier, user.SubscriptionStatus,
		user.MonthlyBgRemovalsLimit, user.MonthlyStorageLimitMB,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return utils.ErrConflict("Email already exists", ErrUserAlreadyExists)
		}
		if err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"` {
			return utils.ErrConflict("Username already exists", ErrUserAlreadyExists)
		}
		return utils.ErrDatabase("Failed to create user", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
        SELECT 
            id, email, username, display_name, password_hash,
            profile_image_url, bio,
            subscription_tier, stripe_customer_id, subscription_status, subscription_expires_at,
            monthly_bg_removals_used, monthly_bg_removals_limit,
            monthly_storage_used_mb, monthly_storage_limit_mb,
            follower_count, following_count, is_verified,
            created_at, updated_at, last_login_at, deleted_at
        FROM auth.users
        WHERE id = $1 AND deleted_at IS NULL
    `

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash,
		&user.ProfileImageURL, &user.Bio,
		&user.SubscriptionTier, &user.StripeCustomerID, &user.SubscriptionStatus, &user.SubscriptionExpiresAt,
		&user.MonthlyBgRemovalsUsed, &user.MonthlyBgRemovalsLimit,
		&user.MonthlyStorageUsedMB, &user.MonthlyStorageLimitMB,
		&user.FollowerCount, &user.FollowingCount, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound("User")
		}
		return nil, utils.ErrDatabase("Failed to get user", err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT 
            id, email, username, display_name, password_hash,
            profile_image_url, bio,
            subscription_tier, stripe_customer_id, subscription_status, subscription_expires_at,
            monthly_bg_removals_used, monthly_bg_removals_limit,
            monthly_storage_used_mb, monthly_storage_limit_mb,
            follower_count, following_count, is_verified,
            created_at, updated_at, last_login_at, deleted_at
        FROM auth.users
        WHERE email = $1 AND deleted_at IS NULL
    `

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash,
		&user.ProfileImageURL, &user.Bio,
		&user.SubscriptionTier, &user.StripeCustomerID, &user.SubscriptionStatus, &user.SubscriptionExpiresAt,
		&user.MonthlyBgRemovalsUsed, &user.MonthlyBgRemovalsLimit,
		&user.MonthlyStorageUsedMB, &user.MonthlyStorageLimitMB,
		&user.FollowerCount, &user.FollowingCount, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound("User")
		}
		return nil, utils.ErrDatabase("Failed to get user", err)
	}

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT 
            id, email, username, display_name, password_hash,
            profile_image_url, bio,
            subscription_tier, stripe_customer_id, subscription_status, subscription_expires_at,
            monthly_bg_removals_used, monthly_bg_removals_limit,
            monthly_storage_used_mb, monthly_storage_limit_mb,
            follower_count, following_count, is_verified,
            created_at, updated_at, last_login_at, deleted_at
        FROM auth.users
        WHERE username = $1 AND deleted_at IS NULL
    `

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash,
		&user.ProfileImageURL, &user.Bio,
		&user.SubscriptionTier, &user.StripeCustomerID, &user.SubscriptionStatus, &user.SubscriptionExpiresAt,
		&user.MonthlyBgRemovalsUsed, &user.MonthlyBgRemovalsLimit,
		&user.MonthlyStorageUsedMB, &user.MonthlyStorageLimitMB,
		&user.FollowerCount, &user.FollowingCount, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound("User")
		}
		return nil, utils.ErrDatabase("Failed to get user", err)
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE auth.users
        SET 
            display_name = $2,
            profile_image_url = $3,
            bio = $4,
            subscription_tier = $5,
            subscription_status = $6,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
        RETURNING updated_at
    `

	err := r.db.Pool.QueryRow(
		ctx, query,
		user.ID, user.DisplayName, user.ProfileImageURL, user.Bio,
		user.SubscriptionTier, user.SubscriptionStatus,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.ErrNotFound("User")
		}
		return utils.ErrDatabase("Failed to update user", err)
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
        UPDATE auth.users
        SET last_login_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
    `

	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		return utils.ErrDatabase("Failed to update last login", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
        UPDATE auth.users
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
    `

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return utils.ErrDatabase("Failed to delete user", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrNotFound("User")
	}

	return nil
}

// Refresh token methods

func (r *userRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
        INSERT INTO auth.refresh_tokens (id, user_id, token_hash, expires_at)
        VALUES ($1, $2, $3, $4)
        RETURNING created_at
    `

	err := r.db.Pool.QueryRow(
		ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		return utils.ErrDatabase("Failed to create refresh token", err)
	}

	return nil
}

func (r *userRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
        SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
        FROM auth.refresh_tokens
        WHERE token_hash = $1 AND revoked_at IS NULL
    `

	token := &models.RefreshToken{}
	err := r.db.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash,
		&token.ExpiresAt, &token.CreatedAt, &token.RevokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrUnauthorized("Invalid refresh token", nil)
		}
		return nil, utils.ErrDatabase("Failed to get refresh token", err)
	}

	return token, nil
}

func (r *userRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
        UPDATE auth.refresh_tokens
        SET revoked_at = CURRENT_TIMESTAMP
        WHERE token_hash = $1 AND revoked_at IS NULL
    `

	_, err := r.db.Pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return utils.ErrDatabase("Failed to revoke refresh token", err)
	}

	return nil
}

func (r *userRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `
        UPDATE auth.refresh_tokens
        SET revoked_at = CURRENT_TIMESTAMP
        WHERE user_id = $1 AND revoked_at IS NULL
    `

	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		return utils.ErrDatabase("Failed to revoke user tokens", err)
	}

	return nil
}
