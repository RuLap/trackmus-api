package user

import (
	"log/slog"

	"github.com/RuLap/trackmus-api/internal/pkg/storage/minio"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	repo    Repository
	service Service
	Handler Handler
}

func NewModule(log *slog.Logger, pool *pgxpool.Pool, minio *minio.Service) *Module {
	repo := NewRepository(pool)

	service := NewService(log, minio, repo)

	handler := NewHandler(log, service)

	return &Module{
		repo:    repo,
		service: service,
		Handler: *handler,
	}
}
