package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/A4AD-team/profile-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("profile not found")

type ProfileRepository interface {
	Create(ctx context.Context, p *domain.Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	GetByUsername(ctx context.Context, username string) (*domain.Profile, error)
	Update(ctx context.Context, userID uuid.UUID, req *domain.UpdateProfileRequest) (*domain.Profile, error)
	IncrementStats(ctx context.Context, userID uuid.UUID, postDelta, commentDelta int) error
	IncrementReputation(ctx context.Context, userID uuid.UUID, delta int) error
	UpdateReputationByID(ctx context.Context, id uuid.UUID, delta int) error
	Search(ctx context.Context, query string, limit, offset int) ([]*domain.Profile, error)
	ListAll(ctx context.Context, limit, offset int) ([]*domain.Profile, error)
}

type postgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) ProfileRepository {
	return &postgresRepo{pool: pool}
}

func (r *postgresRepo) Create(ctx context.Context, p *domain.Profile) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO profiles (id, user_id, username, email, full_name, is_public, joined_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.UserID, p.Username, p.Email, p.FullName, p.IsPublic, p.JoinedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	return nil
}

func (r *postgresRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	p := &domain.Profile{}
	err := r.pool.QueryRow(ctx, `
        SELECT id, user_id, username, email, full_name, bio, location, is_public,
               post_count, comment_count, reputation, joined_at, updated_at
        FROM profiles WHERE user_id = $1`, userID,
	).Scan(
		&p.ID, &p.UserID, &p.Username, &p.Email, &p.FullName, &p.Bio, &p.Location,
		&p.IsPublic, &p.PostCount, &p.CommentCount, &p.Reputation,
		&p.JoinedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get profile by user_id: %w", err)
	}
	return p, nil
}

func (r *postgresRepo) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	p := &domain.Profile{}
	err := r.pool.QueryRow(ctx, `
        SELECT id, user_id, username, email, full_name, bio, location, is_public,
               post_count, comment_count, reputation, joined_at, updated_at
        FROM profiles WHERE username = $1`, username,
	).Scan(
		&p.ID, &p.UserID, &p.Username, &p.Email, &p.FullName, &p.Bio, &p.Location,
		&p.IsPublic, &p.PostCount, &p.CommentCount, &p.Reputation,
		&p.JoinedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get profile by username: %w", err)
	}
	return p, nil
}

func (r *postgresRepo) Update(ctx context.Context, userID uuid.UUID, req *domain.UpdateProfileRequest) (*domain.Profile, error) {
	p := &domain.Profile{}
	err := r.pool.QueryRow(ctx, `
        UPDATE profiles SET
            full_name  = COALESCE($1, full_name),
            bio        = COALESCE($2, bio),
            location   = COALESCE($3, location),
            is_public  = COALESCE($4, is_public),
            updated_at = NOW()
        WHERE user_id = $5
        RETURNING id, user_id, username, email, full_name, bio, location, is_public,
                  post_count, comment_count, reputation, joined_at, updated_at`,
		req.FullName, req.Bio, req.Location, req.IsPublic, userID,
	).Scan(
		&p.ID, &p.UserID, &p.Username, &p.Email, &p.FullName, &p.Bio, &p.Location,
		&p.IsPublic, &p.PostCount, &p.CommentCount, &p.Reputation,
		&p.JoinedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return p, nil
}

func (r *postgresRepo) IncrementStats(ctx context.Context, userID uuid.UUID, postDelta, commentDelta int) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE profiles SET
            post_count    = post_count + $1,
            comment_count = comment_count + $2,
            updated_at    = NOW()
        WHERE user_id = $3`,
		postDelta, commentDelta, userID,
	)
	if err != nil {
		return fmt.Errorf("increment stats: %w", err)
	}
	return nil
}

func (r *postgresRepo) IncrementReputation(ctx context.Context, userID uuid.UUID, delta int) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE profiles SET
            reputation = reputation + $1,
            updated_at = NOW()
        WHERE user_id = $2`,
		delta, userID,
	)
	if err != nil {
		return fmt.Errorf("increment reputation by user_id: %w", err)
	}
	return nil
}

func (r *postgresRepo) UpdateReputationByID(ctx context.Context, id uuid.UUID, delta int) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE profiles SET
            reputation = reputation + $1,
            updated_at = NOW()
        WHERE id = $2`,
		delta, id,
	)
	if err != nil {
		return fmt.Errorf("update reputation by id: %w", err)
	}
	return nil
}

func (r *postgresRepo) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Profile, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, user_id, username, email, full_name, bio, location, is_public,
               post_count, comment_count, reputation, joined_at, updated_at
        FROM profiles
        WHERE username ILIKE '%' || $1 || '%'
           OR bio      ILIKE '%' || $1 || '%'
        ORDER BY reputation DESC
        LIMIT $2 OFFSET $3`,
		query, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("search profiles: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *postgresRepo) ListAll(ctx context.Context, limit, offset int) ([]*domain.Profile, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, user_id, username, email, full_name, bio, location, is_public,
               post_count, comment_count, reputation, joined_at, updated_at
        FROM profiles
        ORDER BY joined_at DESC
        LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list all profiles: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func scanRows(rows pgx.Rows) ([]*domain.Profile, error) {
	var profiles []*domain.Profile
	for rows.Next() {
		p := &domain.Profile{}
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Username, &p.Email, &p.FullName, &p.Bio, &p.Location,
			&p.IsPublic, &p.PostCount, &p.CommentCount, &p.Reputation,
			&p.JoinedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}
