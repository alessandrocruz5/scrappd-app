package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	Username        string    `json:"username"`
	DisplayName     *string   `json:"display_name,omitempty"`
	PasswordHash    string    `json:"-"` // Never expose in JSON
	ProfileImageURL *string   `json:"profile_image_url,omitempty"`
	Bio             *string   `json:"bio,omitempty"`

	// Subscription
	SubscriptionTier      string     `json:"subscription_tier"`
	StripeCustomerID      *string    `json:"-"`
	SubscriptionStatus    string     `json:"subscription_status"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`

	// Usage limits
	MonthlyBgRemovalsUsed  int     `json:"monthly_bg_removals_used"`
	MonthlyBgRemovalsLimit int     `json:"monthly_bg_removals_limit"`
	MonthlyStorageUsedMB   float64 `json:"monthly_storage_used_mb"`
	MonthlyStorageLimitMB  int     `json:"monthly_storage_limit_mb"`

	// Social stats
	FollowerCount  int  `json:"follower_count"`
	FollowingCount int  `json:"following_count"`
	IsVerified     bool `json:"is_verified"`

	// Metadata
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	DeletedAt   *time.Time `json:"-"`
}

type CreateUserRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Username    string  `json:"username" binding:"required,min=3,max=50"`
	Password    string  `json:"password" binding:"required,min=8"`
	DisplayName *string `json:"display_name,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}
