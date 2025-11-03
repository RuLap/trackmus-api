package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository interface {
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	Create(ctx context.Context, session *Session, taskID uuid.UUID) (*Session, error)
}

type sessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) SessionRepository {
	return &sessionRepository{pool}
}

func (r *sessionRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]Session, error) {
	query := `
		SELECT id, bpm, note, confidence, start_time, end_time
		FROM sessions
		WHERE task_id = $1
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	sessions := make([]Session, 0)
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.BPM,
			&session.Note,
			&session.Confidence,
			&session.StartTime,
			&session.EndTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Session, error) {
	query := `
		SELECT id, bpm, note, confidence, start_time, end_time
		FROM sessions
		WHERE id = $1
	`

	var session Session
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&session.ID,
		&session.BPM,
		&session.Note,
		&session.Confidence,
		&session.StartTime,
		&session.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepository) Create(ctx context.Context, session *Session, taskID uuid.UUID) (*Session, error) {
	query := `
		INSERT INTO sessions(task_id, bpm, note, confidence, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx,
		query,
		taskID,
		session.BPM,
		session.Note,
		session.Confidence,
		session.StartTime,
		session.EndTime,
	).Scan(
		&id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	session.ID = id

	return session, nil
}
