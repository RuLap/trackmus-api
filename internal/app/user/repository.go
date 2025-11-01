package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, model *User) (*User, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool}
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, first_name, last_name, username
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *repository) Update(ctx context.Context, model *User) (*User, error) {
	query := `
		UPDATE users
		SET first_name = $2
			last_name = $3
			username = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		model.ID,
		model.FirstName,
		model.LastName,
		model.Username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return model, nil
}
