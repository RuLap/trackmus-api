package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository interface {
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]Media, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	Create(ctx context.Context, model *Media) (*Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByIDs(ctx context.Context, ids []uuid.UUID) error
}

type mediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) MediaRepository {
	return &mediaRepository{pool}
}

func (r *mediaRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]Media, error) {
	query := `
		SELECT id, type, filename, url, size, duration, created_at
		FROM media
		WHERE task_id = $1
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {

	}
	defer rows.Close()

	var medias []Media
	for rows.Next() {
		var media Media
		err := rows.Scan(
			&media.ID,
			&media.Type,
			&media.Filename,
			&media.Size,
			&media.Duration,
			&media.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}

		medias = append(medias, media)
	}

	return medias, nil
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*Media, error) {
	query := `
		SELECT id, type, filename, url, size, duration, created_at
		FROM media
		WHERE id = $1
	`

	var media Media
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&media.ID,
		&media.Type,
		&media.Filename,
		&media.Size,
		&media.Duration,
		&media.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan media: %w", err)
	}

	return &media, nil
}

func (r *mediaRepository) Create(ctx context.Context, model *Media) (*Media, error) {
	query := `
		INSERT INTO media(type, filename, url, size, duration)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx,
		query,
		model.Type,
		model.Filename,
		model.Size,
		model.Duration,
	).Scan(
		&id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}
	model.ID = id

	return model, nil
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM media
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	return nil
}

func (r *mediaRepository) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	query := `
		DELETE FROM media
		WHERE id = ANY($1::uuid[])
	`

	_, err := r.pool.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	return nil
}
