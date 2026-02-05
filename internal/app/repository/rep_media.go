package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/w1zZzyy22/art-analysis/internal/app/model"
	"gorm.io/gorm"
)

// GetExpertWithMedia возвращает эксперта вместе с его медиафайлами
func (r *Repository) GetExpertWithMedia(id uint) (*model.ArtExpert, error) {
	var expert model.ArtExpert
	err := r.db.Preload("Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("id_media ASC")
	}).First(&expert, id).Error
	if err != nil {
		return nil, err
	}
	return &expert, nil
}

// GetMediaByExpertID возвращает все медиафайлы для эксперта, отсортированные по id_media
func (r *Repository) GetMediaByExpertID(expertID uint) ([]model.ExpertMedia, error) {
	var media []model.ExpertMedia
	err := r.db.Where("id_artcenter = ?", expertID).Order("id_media ASC").Find(&media).Error
	if err != nil {
		return nil, err
	}
	return media, nil
}

// GetMediaByID возвращает медиафайл по его ID
func (r *Repository) GetMediaByID(mediaID uint) (*model.ExpertMedia, error) {
	var media model.ExpertMedia
	err := r.db.First(&media, mediaID).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// AddExpertMedia загружает медиафайл в MinIO и сохраняет запись в БД
func (r *Repository) AddExpertMedia(ctx context.Context, expertID uint, fileHeader *multipart.FileHeader) (*model.ExpertMedia, error) {
	// Проверяем существование эксперта
	var expert model.ArtExpert
	if err := r.db.First(&expert, expertID).Error; err != nil {
		return nil, fmt.Errorf("expert not found: %w", err)
	}

	// Определяем тип медиа по расширению/MIME-type
	contentType := fileHeader.Header.Get("Content-Type")
	mediaType := determineMediaType(contentType, fileHeader.Filename)

	// Формируем имя файла
	fileName := filepath.Base(fileHeader.Filename)
	ext := filepath.Ext(fileName)
	name := fileName[:len(fileName)-len(ext)]
	if name == "" {
		name = "media"
	}
	safeName := strings.ReplaceAll(name, " ", "-")
	// Добавляем timestamp для уникальности
	timestamp := fmt.Sprintf("%d", ctx.Value("request_id"))
	if timestamp == "<nil>" {
		timestamp = fmt.Sprintf("%d", expertID)
	}
	fileName = fmt.Sprintf("%s_%d%s", safeName, expertID, ext)
	objectName := "media/" + fileName

	// Открываем файл
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Загружаем в MinIO
	if r.minio == nil {
		return nil, fmt.Errorf("object storage is not configured")
	}
	if err := r.minio.ensureBucket(ctx); err != nil {
		return nil, err
	}

	_, err = r.minio.client.PutObject(ctx, r.minio.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, err
	}

	// Формируем URL
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "127.0.0.1:9000"
	}
	mediaURL := fmt.Sprintf("http://%s/%s/%s", minioEndpoint, r.minio.bucketName, objectName)

	// Создаем запись в БД
	media := model.ExpertMedia{
		ID_artcenter: expertID,
		MediaURL:     mediaURL,
		MediaType:    mediaType,
	}

	if err := r.db.Create(&media).Error; err != nil {
		// Если не удалось создать запись, удаляем файл из MinIO
		r.minio.client.RemoveObject(ctx, r.minio.bucketName, objectName, minio.RemoveObjectOptions{})
		return nil, err
	}

	return &media, nil
}

// DeleteExpertMedia удаляет медиафайл по ID (из MinIO и БД)
func (r *Repository) DeleteExpertMedia(ctx context.Context, mediaID uint) error {
	var media model.ExpertMedia
	if err := r.db.First(&media, mediaID).Error; err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Удаляем из MinIO
	if r.minio != nil && media.MediaURL != "" {
		mediaURL, err := url.Parse(media.MediaURL)
		if err == nil {
			objectName := strings.TrimPrefix(mediaURL.Path, fmt.Sprintf("/%s/", r.minio.bucketName))
			r.minio.client.RemoveObject(ctx, r.minio.bucketName, objectName, minio.RemoveObjectOptions{})
		}
	}

	// Удаляем из БД
	return r.db.Delete(&model.ExpertMedia{}, mediaID).Error
}

// DeleteAllExpertMedia удаляет все медиафайлы эксперта
func (r *Repository) DeleteAllExpertMedia(ctx context.Context, expertID uint) error {
	media, err := r.GetMediaByExpertID(expertID)
	if err != nil {
		return err
	}

	for _, m := range media {
		if r.minio != nil && m.MediaURL != "" {
			mediaURL, err := url.Parse(m.MediaURL)
			if err == nil {
				objectName := strings.TrimPrefix(mediaURL.Path, fmt.Sprintf("/%s/", r.minio.bucketName))
				r.minio.client.RemoveObject(ctx, r.minio.bucketName, objectName, minio.RemoveObjectOptions{})
			}
		}
	}

	return r.db.Where("id_artcenter = ?", expertID).Delete(&model.ExpertMedia{}).Error
}

// determineMediaType определяет тип медиа по MIME-type или расширению файла
func determineMediaType(contentType, filename string) string {
	// Проверяем MIME-type
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}

	// Fallback на расширение файла
	ext := strings.ToLower(filepath.Ext(filename))
	videoExtensions := map[string]bool{
		".mp4": true, ".webm": true, ".ogg": true, ".mov": true,
		".avi": true, ".mkv": true, ".m4v": true,
	}
	if videoExtensions[ext] {
		return "video"
	}

	return "image"
}
