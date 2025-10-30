package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepository interface {
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]Link, error)
	Create(ctx context.Context, model *Link) (*Link, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type linkRepository struct {
	pool *pgxpool.Pool
}

func NewLinkRepository(pool *pgxpool.Pool) LinkRepository {
	return &linkRepository{pool}
}

func (r *linkRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]Link, error) {
	query := `
		SELECT id, task_id, url, title, type, created_at
		FROM links
		WHERE task_id = $1
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {

	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		err := rows.Scan(
			&link.ID,
			&link.TaskID,
			&link.Title,
			&link.Type,
			&link.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		links = append(links, link)
	}

	return links, nil
}

func (r *linkRepository) Create(ctx context.Context, model *Link) (*Link, error) {
	query := `
		INSERT INTO links(task_id, url, title, type)
		VALUES($1, $2, $3, $4)
		RETURNING id
	`

	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx,
		query,
		model.TaskID,
		model.Title,
		model.Type,
	).Scan(
		&id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}
	model.ID = id

	return model, nil
}

func (r *linkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM links
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}
