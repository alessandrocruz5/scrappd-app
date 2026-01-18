package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a scrapbook project owned by a user.
type Project struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	CoverImageURL *string    `json:"cover_image_url,omitempty"`
	Visibility    string     `json:"visibility"`
	IsTemplate    bool       `json:"is_template"`
	TemplatePrice *float64   `json:"template_price,omitempty"`
	ViewCount     int        `json:"view_count"`
	LikeCount     int        `json:"like_count"`
	ForkCount     int        `json:"fork_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}

// ProjectUpdate holds optional fields to update.
type ProjectUpdate struct {
	ID            uuid.UUID `json:"-"`
	Title         *string   `json:"title,omitempty"`
	Description   *string   `json:"description,omitempty"`
	CoverImageURL *string   `json:"cover_image_url,omitempty"`
	Visibility    *string   `json:"visibility,omitempty"`
	IsTemplate    *bool     `json:"is_template,omitempty"`
	TemplatePrice *float64  `json:"template_price,omitempty"`
}
