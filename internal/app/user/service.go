package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/RuLap/trackmus-api/internal/pkg/errors"
	"github.com/RuLap/trackmus-api/internal/pkg/storage/minio"
	"github.com/google/uuid"
)

type Service interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*GetUserResponse, error)
	UpdateUser(ctx context.Context, req *SaveUserRequest, id uuid.UUID) (*GetUserResponse, error)
	GetAvatarUploadURL(ctx context.Context, userID uuid.UUID) (*GetUploadURLResponse, error)
	ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID) (*ConfirmUploadAvatarResponse, error)
}

type service struct {
	log        *slog.Logger
	minio      *minio.Service
	bucketName string
	repo       Repository
}

func NewService(log *slog.Logger, minio *minio.Service, repo Repository) Service {
	return &service{
		log:        log,
		minio:      minio,
		bucketName: "trackmus_avatars",
		repo:       repo,
	}
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*GetUserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get user from repository",
			"id", id,
			"isCompleted", false,
			"error", err,
		)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	avatarURL, err := s.getDownloadAvatarURL(ctx, id)
	if err != nil {
		s.log.Error("failed to get download url", "userID", id, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := UserToGetResponse(user, avatarURL)

	return &result, nil
}

func (s *service) UpdateUser(ctx context.Context, req *SaveUserRequest, id uuid.UUID) (*GetUserResponse, error) {
	model := SaveRequestToUser(req)

	user, err := s.repo.Update(ctx, &model)
	if err != nil {
		s.log.Error("failed to update user in repository", "req", req, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	avatarURL, err := s.getDownloadAvatarURL(ctx, id)
	if err != nil {
		s.log.Error("failed to get download url", "userID", id, "error", err)
		return nil, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	result := UserToGetResponse(user, avatarURL)

	return &result, nil
}

func (s *service) GetAvatarUploadURL(ctx context.Context, userID uuid.UUID) (*GetUploadURLResponse, error) {
	s3key := userID.String()

	avatarURL, err := s.minio.GenerateUploadURL(ctx, s.bucketName, s3key)
	if err != nil {
		s.log.Error("failed to generate upload url", "objName", s3key)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	return &GetUploadURLResponse{
		URL: avatarURL,
	}, nil
}

func (s *service) ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID) (*ConfirmUploadAvatarResponse, error) {
	s3key := userID.String()

	avatarURL, err := s.minio.GenerateUploadURL(ctx, s.bucketName, s3key)
	if err != nil {
		s.log.Error("failed to generate upload url", "objName", s3key)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	return &ConfirmUploadAvatarResponse{
		URL: avatarURL,
	}, nil
}

func (s *service) getDownloadAvatarURL(ctx context.Context, userID uuid.UUID) (string, error) {
	s3key := userID.String()
	downloadUrl, err := s.minio.GenerateDownloadURL(ctx, s.bucketName, s3key, "avatar")
	if err != nil {
		s.log.Error("failed to generate download URL",
			"objName", s3key,
			"error", err,
		)
		return "", fmt.Errorf(errors.ErrFailedToLoadData)
	}

	return downloadUrl, nil
}
