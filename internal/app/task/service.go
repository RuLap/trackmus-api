package task

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/RuLap/trackmus-api/internal/pkg/errors"
	"github.com/RuLap/trackmus-api/internal/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	GetActiveTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error)
	GetCompletedTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*GetTaskResponse, error)
	CreateTask(ctx context.Context, req *SaveTaskRequest, userID uuid.UUID) (*GetTaskShortResponse, error)
	UpdateTask(ctx context.Context, req *SaveTaskRequest, id uuid.UUID) (*GetTaskShortResponse, error)
	CompleteTask(ctx context.Context, id uuid.UUID) error

	GetSessionByID(ctx context.Context, id uuid.UUID) (*GetSessionResponse, error)
	CreateSession(ctx context.Context, req *SaveSessionRequest, taskID uuid.UUID) (*GetSessionResponse, error)

	SaveMedia(ctx context.Context, req *SaveMediaRequest, taskID uuid.UUID) (*GetMediaResponse, error)
	RemoveMedia(ctx context.Context, id uuid.UUID) error

	SaveLink(ctx context.Context, req *SaveLinkRequest, taskID uuid.UUID) (*GetLinkResponse, error)
	RemoveLink(ctx context.Context, id uuid.UUID) error
}

type service struct {
	log         *slog.Logger
	redis       *redis.Client
	rabbitmq    *rabbitmq.Client
	taskRepo    TaskRepository
	sessionRepo SessionRepository
	mediaRepo   MediaRepository
	linkRepo    LinkRepository
}

func NewService(
	log *slog.Logger,
	redis *redis.Client,
	rabbitmq *rabbitmq.Client,
	taskRepo TaskRepository,
	sessionRepo SessionRepository,
	mediaRepo MediaRepository,
	linkRepo LinkRepository,
) Service {
	return &service{
		log:         log,
		redis:       redis,
		rabbitmq:    rabbitmq,
		taskRepo:    taskRepo,
		sessionRepo: sessionRepo,
		mediaRepo:   mediaRepo,
		linkRepo:    linkRepo,
	}
}

func (s *service) GetActiveTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error) {
	tasks, err := s.taskRepo.Get(ctx, userID, false)
	if err != nil {
		s.log.Error("failed to get active tasks from repository",
			"error", err,
			"userID", userID,
			"isActive", false,
		)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	var result []GetTaskShortResponse
	for _, task := range tasks {
		progress, err := getTaskProgress()
		if err != nil {

		}
		dto := TaskToGetShortResponse(&task, progress)

		result = append(result, dto)
	}

	return result, nil
}

func (s *service) CompleteTask(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) CreateSession(ctx context.Context, req *SaveSessionRequest, taskID uuid.UUID) (*GetSessionResponse, error) {
	panic("unimplemented")
}

func (s *service) CreateTask(ctx context.Context, req *SaveTaskRequest, userID uuid.UUID) (*GetTaskShortResponse, error) {
	panic("unimplemented")
}

func (s *service) GetCompletedTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error) {
	panic("unimplemented")
}

func (s *service) GetSessionByID(ctx context.Context, id uuid.UUID) (*GetSessionResponse, error) {
	panic("unimplemented")
}

func (s *service) GetTaskByID(ctx context.Context, id uuid.UUID) (*GetTaskResponse, error) {
	panic("unimplemented")
}

func (s *service) RemoveLink(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) RemoveMedia(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) SaveLink(ctx context.Context, req *SaveLinkRequest, taskID uuid.UUID) (*GetLinkResponse, error) {
	panic("unimplemented")
}

func (s *service) SaveMedia(ctx context.Context, req *SaveMediaRequest, taskID uuid.UUID) (*GetMediaResponse, error) {
	panic("unimplemented")
}

func (s *service) UpdateTask(ctx context.Context, req *SaveTaskRequest, id uuid.UUID) (*GetTaskShortResponse, error) {
	panic("unimplemented")
}

func getTaskProgress() (float64, error) {
	return 0, nil
}
