package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	UserID       uuid.UUID `db:"user_id"        json:"userId"`
	Username     string    `db:"username"       json:"username"`
	Email        string    `db:"email"          json:"email"`
	FullName     *string   `db:"full_name"      json:"fullName,omitempty"`
	Bio          *string   `db:"bio"            json:"bio,omitempty"`
	Location     *string   `db:"location"       json:"location,omitempty"`
	IsPublic     bool      `db:"is_public"      json:"isPublic"`
	PostCount    int       `db:"post_count"     json:"postCount"`
	CommentCount int       `db:"comment_count"  json:"commentCount"`
	Reputation   int       `db:"reputation"     json:"reputation"`
	JoinedAt     time.Time `db:"joined_at"      json:"joinedAt"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updatedAt"`
}

type UpdateProfileRequest struct {
	FullName *string `json:"fullName" validate:"omitempty,max=190"`
	Bio      *string `json:"bio"      validate:"omitempty,max=500"`
	Location *string `json:"location" validate:"omitempty,max=100"`
	IsPublic *bool   `json:"isPublic"`
}

// UpdateReputationRequest is used by PATCH /api/v1/admin/profiles/:id/reputation
type UpdateReputationRequest struct {
	Delta int `json:"delta" validate:"required"`
}

type Stats struct {
	PostCount    int `json:"postCount"`
	CommentCount int `json:"commentCount"`
	Reputation   int `json:"reputation"`
}

func (p *Profile) GetStats() Stats {
	return Stats{
		PostCount:    p.PostCount,
		CommentCount: p.CommentCount,
		Reputation:   p.Reputation,
	}
}
