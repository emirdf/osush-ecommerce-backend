package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClient(
	endpoint string,
	accessKey string,
	secretKey string,
	useSSL bool,
) (*minio.Client, error) {

	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			accessKey,
			secretKey,
			"",
		),
		Secure: useSSL,
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}

func CheckBucket(
	ctx context.Context,
	client *minio.Client,
	bucket string,
) error {

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", bucket)
	}

	return nil
}
