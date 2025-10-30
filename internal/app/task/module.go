package task

import (
	"log/slog"

	"github.com/RuLap/trackmus-api/internal/pkg/storage/minio"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	taskRepo    TaskRepository
	sessionRepo SessionRepository
	mediaRepo   MediaRepository
	linkRepo    LinkRepository
	service     Service
	Handler     Handler
}

func NewModule(log *slog.Logger, pool *pgxpool.Pool, minio *minio.Service) *Module {
	taskRepo := NewTaskRepository(pool)
	sessionRepo := NewSessionRepository(pool)
	mediaRepo := NewMediaRepository(pool)
	linkRepo := NewLinkRepository(pool)

	service := NewService(log, minio, taskRepo, sessionRepo, mediaRepo, linkRepo)

	handler := NewHandler(log, service)

	return &Module{
		taskRepo:    taskRepo,
		sessionRepo: sessionRepo,
		mediaRepo:   mediaRepo,
		linkRepo:    linkRepo,
		service:     service,
		Handler:     *handler,
	}
}
