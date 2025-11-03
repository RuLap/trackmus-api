package task

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/RuLap/trackmus-api/internal/pkg/errors"
	"github.com/RuLap/trackmus-api/internal/pkg/storage/minio"
	"github.com/google/uuid"
)

type Service interface {
	GetActiveTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error)
	GetCompletedTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*GetTaskResponse, error)
	CreateTask(ctx context.Context, req *SaveTaskRequest, userID uuid.UUID) (*GetTaskShortResponse, error)
	UpdateTask(ctx context.Context, req *SaveTaskRequest, id uuid.UUID) (*GetTaskResponse, error)
	CompleteTask(ctx context.Context, id uuid.UUID) (*GetTaskShortResponse, error)

	GetSessionByID(ctx context.Context, id uuid.UUID) (*GetSessionResponse, error)
	CreateSession(ctx context.Context, req *SaveSessionRequest, taskID uuid.UUID) (*GetSessionResponse, error)

	GetMediaUploadURL(ctx context.Context, taskID, mediaID uuid.UUID) (*GetUploadURLResponse, error)
	ConfirmMediaUpload(ctx context.Context, req *ConfirmMediaUploadRequest, id uuid.UUID) (*GetMediaResponse, error)
	RemoveMedia(ctx context.Context, id uuid.UUID) error

	SaveLink(ctx context.Context, req *SaveLinkRequest, taskID uuid.UUID) (*GetLinkResponse, error)
	RemoveLink(ctx context.Context, id uuid.UUID) error
}

type service struct {
	log         *slog.Logger
	minio       *minio.Service
	bucketName  string
	taskRepo    TaskRepository
	sessionRepo SessionRepository
	mediaRepo   MediaRepository
	linkRepo    LinkRepository
}

func NewService(
	log *slog.Logger,
	minio *minio.Service,
	taskRepo TaskRepository,
	sessionRepo SessionRepository,
	mediaRepo MediaRepository,
	linkRepo LinkRepository,
) Service {
	return &service{
		log:         log,
		taskRepo:    taskRepo,
		minio:       minio,
		bucketName:  "trackmus",
		sessionRepo: sessionRepo,
		mediaRepo:   mediaRepo,
		linkRepo:    linkRepo,
	}
}

func (s *service) GetActiveTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error) {
	tasks, err := s.taskRepo.Get(ctx, userID, false)
	if err != nil {
		s.log.Error("failed to get active tasks from repository",
			"userID", userID,
			"isCompleted", false,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := make([]GetTaskShortResponse, 0)
	for _, task := range tasks {
		progress, err := s.getTaskProgress(ctx, &task)
		if err != nil {

		}
		dto := TaskToGetShortResponse(&task, progress)

		result = append(result, dto)
	}

	return result, nil
}

func (s *service) GetCompletedTasks(ctx context.Context, userID uuid.UUID) ([]GetTaskShortResponse, error) {
	tasks, err := s.taskRepo.Get(ctx, userID, true)
	if err != nil {
		s.log.Error("failed to get completed tasks from repository",
			"userID", userID,
			"isCompleted", true,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := make([]GetTaskShortResponse, 0)
	for _, task := range tasks {
		progress, err := s.getTaskProgress(ctx, &task)
		if err != nil {

		}
		dto := TaskToGetShortResponse(&task, progress)

		result = append(result, dto)
	}

	return result, nil
}

func (s *service) GetTaskByID(ctx context.Context, id uuid.UUID) (*GetTaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get task from repository", "id", id, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	sessions, err := s.getSessionsByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	media, err := s.getMediaByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	links, err := s.getLinksByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := TaskToGetResponse(task, sessions, media, links)

	return &result, nil
}

func (s *service) CreateTask(ctx context.Context, req *SaveTaskRequest, userID uuid.UUID) (*GetTaskShortResponse, error) {
	model := SaveRequestToTask(req, userID)

	task, err := s.taskRepo.Create(ctx, &model)
	if err != nil {
		s.log.Error("failed to create task from repository",
			"req", req,
			"userID", userID,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	result := TaskToGetShortResponse(task, 0)

	return &result, nil
}

func (s *service) UpdateTask(ctx context.Context, req *SaveTaskRequest, id uuid.UUID) (*GetTaskResponse, error) {
	model := SaveRequestToTask(req, id)

	task, err := s.taskRepo.Update(ctx, &model)
	if err != nil {
		s.log.Error("failed to save task in repository",
			"req", req,
			"id", id,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	sessions, err := s.getSessionsByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	media, err := s.getMediaByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	links, err := s.getLinksByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := TaskToGetResponse(task, sessions, media, links)

	return &result, nil
}

func (s *service) CompleteTask(ctx context.Context, id uuid.UUID) (*GetTaskShortResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get task from repository", "id", id, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	task.IsCompleted = true

	_, err = s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	progress, err := s.getTaskProgress(ctx, task)
	if err != nil {
		return nil, fmt.Errorf(errors.ErrCommon)
	}

	result := TaskToGetShortResponse(task, progress)

	return &result, nil
}

func (s *service) GetSessionByID(ctx context.Context, id uuid.UUID) (*GetSessionResponse, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get session from repository")
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := SessionToGetResponse(session)

	return &result, nil
}

func (s *service) CreateSession(ctx context.Context, req *SaveSessionRequest, taskID uuid.UUID) (*GetSessionResponse, error) {
	model := SaveRequestToSession(req, taskID)

	session, err := s.sessionRepo.Create(ctx, &model)
	if err != nil {
		s.log.Error("failed to create session in repository",
			"req", req,
			"taskID", taskID,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	result := SessionToGetResponse(session)

	return &result, nil
}

func (s *service) GetMediaUploadURL(ctx context.Context, taskID, mediaID uuid.UUID) (*GetUploadURLResponse, error) {
	s3Key := fmt.Sprintf("%s/%s", taskID, mediaID)

	url, err := s.minio.GenerateUploadURL(ctx, s.bucketName, s3Key)
	if err != nil {
		s.log.Error("failed to generate upload url", "objName", s3Key)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	return &GetUploadURLResponse{
		MediaID: mediaID.String(),
		URL:     url,
	}, nil
}

func (s *service) ConfirmMediaUpload(ctx context.Context, req *ConfirmMediaUploadRequest, id uuid.UUID) (*GetMediaResponse, error) {
	model := ConfirmUploadRequestToMedia(req, id)

	media, err := s.mediaRepo.Create(ctx, &model)
	if err != nil {
		s.log.Error("failed to create media in repository",
			"req", req,
			"id", id,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	s3key := fmt.Sprintf("%s/%s", media.TaskID, media.ID)
	downloadURL, err := s.minio.GenerateDownloadURL(ctx, s.bucketName, s3key, req.Filename)
	if err != nil {
		s.log.Error("failed to generate minio download url")
		err := s.mediaRepo.Delete(ctx, media.ID)
		if err != nil {
			s.log.Warn("failed to delete media in repository", "id", media.ID, "error", err)
		}
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	result := MediaToGetResponse(media, downloadURL)

	return &result, nil
}

func (s *service) RemoveMedia(ctx context.Context, id uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get media from repository", "id", id, "error", err)
		return fmt.Errorf(errors.ErrFailedToLoadData)
	}

	s3key := fmt.Sprintf("%s/%s", media.ID, id)

	err = s.minio.DeleteFile(ctx, s.bucketName, s3key)
	if err != nil {
		s.log.Error("failed to delete media in s3 minio", "objName", s3key, "error", err)
	}

	err = s.mediaRepo.Delete(ctx, id)
	if err != nil {
		s.log.Error("failed to delete media in repository", "id", id, "error", err)
		return fmt.Errorf(errors.ErrFailedToDeleteData)
	}

	s.log.Info("media removed successfuly", "id", id, "taskID", media.TaskID)

	return nil
}

func (s *service) SaveLink(ctx context.Context, req *SaveLinkRequest, taskID uuid.UUID) (*GetLinkResponse, error) {
	model := SaveRequestToLink(req, taskID)

	link, err := s.linkRepo.Create(ctx, &model)
	if err != nil {
		s.log.Error("failed to save link in repository",
			"req", req,
			"taskID", taskID,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	dto := LinkToGetResponse(link)

	return &dto, nil
}

func (s *service) RemoveLink(ctx context.Context, id uuid.UUID) error {
	err := s.linkRepo.Delete(ctx, id)
	if err != nil {
		s.log.Error("failed to delete link in repository", "id", id, "error", err)
		return fmt.Errorf(errors.ErrFailedToDeleteData)
	}

	return nil
}

func (s *service) getSessionsByTaskID(ctx context.Context, taskID uuid.UUID) ([]GetSessionResponse, error) {
	sessions, err := s.sessionRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		s.log.Error("failed to get sessions from repository", "taskID", taskID, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := make([]GetSessionResponse, 0)
	for _, s := range sessions {
		dto := SessionToGetResponse(&s)
		result = append(result, dto)
	}

	return result, nil
}

func (s *service) getMediaByTaskID(ctx context.Context, taskID uuid.UUID) ([]GetMediaResponse, error) {
	models, err := s.mediaRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		s.log.Error("failed to get media from repository", "taskID", taskID, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	if len(models) == 0 {
		return make([]GetMediaResponse, 0), nil
	}

	var result []GetMediaResponse
	var toRemove []Media
	for _, m := range models {
		s3Key := fmt.Sprintf("%s/%s", taskID, m.ID)

		exists, err := s.minio.FileExists(ctx, s.bucketName, s3Key)
		if err != nil {
			s.log.Warn("failed to check file existence",
				"fileID", m.ID,
				"objName", s3Key,
				"error", err,
			)
			continue
		}

		if exists {
			downloadUrl, err := s.minio.GenerateDownloadURL(ctx, s.bucketName, s3Key, m.Filename)
			if err != nil {
				s.log.Error("failed to generate download URL",
					"fileID", m.ID,
					"objName", s3Key,
					"error", err,
				)
				return nil, fmt.Errorf(errors.ErrFailedToLoadData)
			}

			dto := MediaToGetResponse(&m, downloadUrl)
			result = append(result, dto)
		} else {
			s.log.Warn("file found in DB but missing in minio", "fileID", m.ID, "taskID", taskID)
			toRemove = append(toRemove, m)
		}
	}

	if len(toRemove) > 0 {
		go s.cleanupOrphanedFiles(ctx, toRemove)
	}

	return result, nil
}

func (s *service) cleanupOrphanedFiles(ctx context.Context, orphanedFiles []Media) {
	var orphanedIDs []uuid.UUID
	for _, file := range orphanedFiles {
		orphanedIDs = append(orphanedIDs, file.ID)
	}

	s.log.Info("cleaning up orphaned media records", "count", len(orphanedIDs))

	if err := s.mediaRepo.DeleteByIDs(ctx, orphanedIDs); err != nil {
		s.log.Error("failed to cleanup orphaned media", "error", err, "count", len(orphanedIDs))
	} else {
		s.log.Info("successfully cleaned up orphaned media", "count", len(orphanedIDs))
	}
}

func (s *service) getLinksByTaskID(ctx context.Context, taskID uuid.UUID) ([]GetLinkResponse, error) {
	links, err := s.linkRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		s.log.Error("failed to get links from repository", "taskID", taskID, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := make([]GetLinkResponse, 0)
	for _, l := range links {
		dto := LinkToGetResponse(&l)
		result = append(result, dto)
	}

	return result, nil
}

func (s *service) getTaskProgress(ctx context.Context, task *Task) (float64, error) {
	sessions, err := s.sessionRepo.GetByTaskID(ctx, task.ID)
	if err != nil {
		s.log.Error("failed to load sessions from repository", "taskID", task.ID, "error", err)
		return 0, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	if len(sessions) == 0 {
		return 0, nil
	}

	const (
		bw = 0.4
		cw = 0.6
	)

	bestWeightedSum := -1.0

	for _, s := range sessions {
		bprogress := tanhProgress(float64(s.BPM), float64(task.TargetBPM))
		cprogress := tanhProgress(float64(s.Confidence), 5.0)

		weightedSum := bw*bprogress + cw*cprogress

		if weightedSum > bestWeightedSum {
			bestWeightedSum = weightedSum
		}
	}

	return math.Min(bestWeightedSum, 100), nil
}

func tanhProgress(current, target float64) float64 {
	if current >= target {
		return 100.0
	}

	ratio := current / target

	k := 2.2
	x0 := 0.72
	progress := (math.Tanh(k*(ratio-x0)) + 1.0) / 2.0 * 100.0

	return math.Max(0.0, progress)
}
