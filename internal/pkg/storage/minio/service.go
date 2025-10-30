package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client}
}

func (s *Service) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.GetClient().BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.GetClient().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

func (s *Service) UploadFile(ctx context.Context, bucketName, objName string, file io.Reader, size int64) error {
	_, err := s.client.GetClient().PutObject(ctx, bucketName, objName, file, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *Service) DownloadFile(ctx context.Context, bucketName, objName string) (*minio.Object, error) {
	object, err := s.client.GetClient().GetObject(ctx, bucketName, objName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

func (s *Service) GetFileURL(bucketName, objName string) string {
	return fmt.Sprintf("http://%s/%s/%s", s.client.config.Endpoint, bucketName, objName)
}

func (s *Service) DeleteFile(ctx context.Context, bucketName, objName string) error {
	err := s.client.GetClient().RemoveObject(ctx, bucketName, objName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *Service) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	buckets, err := s.client.GetClient().ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	return buckets, nil
}

func (s *Service) CreatePresignedURL(ctx context.Context, bucketName, objName string, expirySec int) (string, error) {
	url, err := s.client.GetClient().PresignedGetObject(ctx, bucketName, objName, time.Duration(expirySec)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create presidned URL: %w", err)
	}

	return url.String(), nil
}
