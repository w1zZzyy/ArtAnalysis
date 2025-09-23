package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client
var MinioBucket string
var MinioBaseURL string

func InitMinioClient() {
	endpoint := os.Getenv("MINIO_BASE_URL") // http://minio:9000
	accessKey := os.Getenv("MINIO_ROOT_USER")
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	MinioBucket = os.Getenv("MINIO_BUCKET")

	if endpoint == "" || accessKey == "" || secretKey == "" || MinioBucket == "" {
		log.Fatal("Не заданы переменные окружения для MinIO")
	}

	MinioBaseURL = endpoint
	// Убираем протокол для minio.New
	urlWithoutProto := strings.TrimPrefix(endpoint, "http://")
	urlWithoutProto = strings.TrimPrefix(urlWithoutProto, "https://")

	client, err := minio.New(urlWithoutProto, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: strings.HasPrefix(endpoint, "https://"),
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, MinioBucket)
	if err != nil {
		log.Fatalf("Ошибка проверки bucket %s: %v", MinioBucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, MinioBucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("Ошибка создания bucket %s: %v", MinioBucket, err)
		}
		log.Printf("Bucket %s создан", MinioBucket)
	} else {
		log.Printf("Bucket %s уже существует", MinioBucket)
	}

	MinioClient = client
}

// BuildImageURL формирует публичный URL для объекта
func BuildImageURL(imageKey string) string {
	path := fmt.Sprintf("http://localhost:9000/%s/%s", MinioBucket, imageKey)
	fmt.Println(path)
	return path
}
