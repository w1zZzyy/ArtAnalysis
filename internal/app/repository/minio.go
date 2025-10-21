package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// minioClient wraps the SDK to keep dependency localized in repo layer
type minioClient struct {
	client     *minio.Client
	bucketName string
}

func newMinioClientFromEnv() (*minioClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	useSSL := strings.ToLower(os.Getenv("MINIO_USE_SSL")) == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("minio env vars are not fully set")
	}

	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &minioClient{client: cli, bucketName: bucket}, nil
}

func (mc *minioClient) ensureBucket(ctx context.Context) error {
	exists, err := mc.client.BucketExists(ctx, mc.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return mc.client.MakeBucket(ctx, mc.bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

// UploadImage uploads a file stream to MinIO under img/<fileName>
func (r *Repository) UploadImage(ctx context.Context, file multipart.File, size int64, fileName string, contentType string) (string, error) {
	if r.minio == nil {
		return "", fmt.Errorf("object storage is not configured")
	}
	if err := r.minio.ensureBucket(ctx); err != nil {
		return "", err
	}
	objectName := path.Join("img", fileName)
	_, err := r.minio.client.PutObject(ctx, r.minio.bucketName, objectName, file, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return objectName, nil
}
