package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

type MinioRepository struct {
	client *minio.Client
	bucket string
}

func NewMinioRepository(client *minio.Client, bucket string) *MinioRepository {

	return &MinioRepository{
		client: client,
		bucket: bucket,
	}
}

func (r *MinioRepository) Upload(
	ctx context.Context,
	objectName string,
	file io.Reader,
	size int64,
	contentType string,
) error {

	_, err := r.client.PutObject(
		ctx,
		r.bucket,
		objectName,
		file,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)

	return err
}

func (r *MinioRepository) Delete(ctx context.Context, objectName string) error {
	return r.client.RemoveObject(
		ctx,
		r.bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
}
