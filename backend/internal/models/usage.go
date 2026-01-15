package models

import (
	"time"

	"github.com/google/uuid"
)

// UsageTracking represents a user's usage for a specific period
type UsageTracking struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// Usage period
	PeriodStart time.Time `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time `json:"period_end" db:"period_end"`

	// Counters
	ItemsProcessed int  `json:"items_processed" db:"items_processed"`
	ItemsLimit     *int `json:"items_limit,omitempty" db:"items_limit"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UsageStats struct {
	ItemsProcessed int       `json:"items_processed"`
	ItemsLimt      *int      `json:"items_limit,omitempty"`
	ItemsRemaining *int      `json:"items_remaining,omitempty"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	IsUnlimited    bool      `json:"is_unlimited"`
}

type SubscriptionTier string

const (
	TierFree SubscriptionTier = "free"
	TierPro  SubscriptionTier = "pro"
)

func (t SubscriptionTier) GetItemsLimit() *int {
	switch t {
	case TierFree:
		limit := 5
		return &limit
	case TierPro:
		return nil
	default:
		limit := 5
		return &limit
	}
}

func (t SubscriptionTier) IsUnlimited() bool {
	return t == TierPro
}
