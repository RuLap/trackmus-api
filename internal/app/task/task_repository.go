package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository interface {
	Get(ctx context.Context, userID uuid.UUID, isCompleted bool) ([]Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	Create(ctx context.Context, task *Task, userID uuid.UUID) (*Task, error)
	Update(ctx context.Context, task *Task) (*Task, error)
}

type taskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &taskRepository{pool}
}

func (r *taskRepository) Get(ctx context.Context, userID uuid.UUID, isCompleted bool) ([]Task, error) {
	query := `
		SELECT id, title, target_bpm, is_completed, created_at
		FROM tasks
		WHERE user_id = $1 AND is_completed = $2
	`

	rows, err := r.pool.Query(ctx, query, userID, isCompleted)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.TargetBPM,
			&task.IsCompleted,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	query := `
		SELECT id, title, target_bpm, is_completed
		FROM tasks
		WHERE id = $1
	`

	var task Task
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&task.ID,
		&task.Title,
		&task.TargetBPM,
		&task.IsCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return &task, nil
}

func (r *taskRepository) Create(ctx context.Context, task *Task, userID uuid.UUID) (*Task, error) {
	query := `
		INSERT INTO tasks(user_id, title, target_bpm)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
		task.Title,
		task.TargetBPM,
	).Scan(
		&id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	task.ID = id

	return task, nil
}

func (r *taskRepository) Update(ctx context.Context, task *Task) (*Task, error) {
	query := `
		UPDATE tasks
		SET title = $2,
			target_bpm = $3,
			is_completed = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		task.ID,
		task.Title,
		task.TargetBPM,
		task.IsCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}
