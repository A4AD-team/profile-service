package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/A4AD-team/profile-service/internal/domain"
	"github.com/A4AD-team/profile-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrAlreadyExists   = errors.New("profile already exists")
)

type ProfileService struct {
	repo repository.ProfileRepository
}

func NewProfileService(repo repository.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

// CreateProfile is called by auth-service via POST /internal/profiles after user registration.
func (s *ProfileService) CreateProfile(ctx context.Context, userID uuid.UUID, username, email string) (*domain.Profile, error) {
	p := &domain.Profile{
		ID:        uuid.New(),
		UserID:    userID,
		Username:  username,
		Email:     email,
		IsPublic:  true,
		JoinedAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, p); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return p, nil
}

// GetMyProfile returns the full profile of the authenticated user.
func (s *ProfileService) GetMyProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

// GetProfileByUsername returns a profile by username.
// All routes require JWT so no need to filter by is_public for now.
func (s *ProfileService) GetProfileByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	p, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

// UpdateProfile updates bio, location and visibility of the authenticated user.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *domain.UpdateProfileRequest) (*domain.Profile, error) {
	p, err := s.repo.Update(ctx, userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

// GetStats returns activity counters for the authenticated user.
func (s *ProfileService) GetStats(ctx context.Context, userID uuid.UUID) (*domain.Stats, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	stats := p.GetStats()
	return &stats, nil
}

// GetProfileByAuthorID returns a profile by author ID (user_id from auth service).
func (s *ProfileService) GetProfileByAuthorID(ctx context.Context, authorID int64) (*domain.Profile, error) {
	p, err := s.repo.GetByAuthorID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

// SearchProfiles searches profiles by username or bio keyword.
func (s *ProfileService) SearchProfiles(ctx context.Context, q string, limit, offset int) ([]*domain.Profile, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.Search(ctx, q, limit, offset)
}

// ListAllProfiles returns all profiles, used by admin endpoint.
func (s *ProfileService) ListAllProfiles(ctx context.Context, limit, offset int) ([]*domain.Profile, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListAll(ctx, limit, offset)
}

// HandlePostCreated increments post_count by 1.
// Called by RabbitMQ consumer on "post.created" event.
func (s *ProfileService) HandlePostCreated(ctx context.Context, authorID uuid.UUID) error {
	return s.repo.IncrementStats(ctx, authorID, +1, 0)
}

// HandlePostDeleted decrements post_count by 1.
// Called by RabbitMQ consumer on "post.deleted" event.
func (s *ProfileService) HandlePostDeleted(ctx context.Context, authorID uuid.UUID) error {
	return s.repo.IncrementStats(ctx, authorID, -1, 0)
}

// HandleCommentCreated increments comment_count by 1.
// Called by RabbitMQ consumer on "comment.created" event.
func (s *ProfileService) HandleCommentCreated(ctx context.Context, authorID uuid.UUID) error {
	return s.repo.IncrementStats(ctx, authorID, 0, +1)
}

// HandleCommentDeleted decrements comment_count by 1.
// Called by RabbitMQ consumer on "comment.deleted" event.
func (s *ProfileService) HandleCommentDeleted(ctx context.Context, authorID uuid.UUID) error {
	return s.repo.IncrementStats(ctx, authorID, 0, -1)
}

// HandlePostLiked updates reputation based on post like/unlike.
// Called by RabbitMQ consumer on "post.liked" event.
func (s *ProfileService) HandlePostLiked(ctx context.Context, authorID uuid.UUID, delta int) error {
	return s.repo.IncrementReputation(ctx, authorID, delta)
}

// HandleCommentLiked updates reputation based on comment like/unlike.
// Called by RabbitMQ consumer on "comment.liked" event.
func (s *ProfileService) HandleCommentLiked(ctx context.Context, authorID uuid.UUID, delta int) error {
	return s.repo.IncrementReputation(ctx, authorID, delta)
}

// UpdateReputation manually adjusts reputation by profile ID, admin only.
func (s *ProfileService) UpdateReputation(ctx context.Context, profileID uuid.UUID, delta int) error {
	return s.repo.UpdateReputationByID(ctx, profileID, delta)
}
