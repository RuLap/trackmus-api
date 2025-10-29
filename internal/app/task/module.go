package task

import (
	"log/slog"

	"github.com/RuLap/trackmus-api/internal/pkg/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	taskRepo    TaskRepository
	sessionRepo SessionRepository
	mediaRepo   MediaRepository
	linkRepo    LinkRepository
	service     Service
	Handler     Handler
}

func NewModule(log *slog.Logger, pool *pgxpool.Pool, redis *redis.Client, rabbitmq *rabbitmq.Client) *Module {
	taskRepo := NewTaskRepository(pool)
	sessionRepo := NewSessionRepository(pool)
	mediaRepo := NewMediaRepository(pool)
	linkRepo := NewLinkRepository(pool)

	service := NewService(log, redis, rabbitmq, taskRepo, sessionRepo, mediaRepo, linkRepo)

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
