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
	"gorm.io/gorm/clause"
)

func (r *Repository) GetExperts() ([]model.ArtExpert, error) {
	var experts []model.ArtExpert

	err := r.db.Find(&experts).Error
	if err != nil {
		return nil, err
	}

	if len(experts) == 0 {
		return nil, fmt.Errorf("experts not found")
	}
	return experts, nil
}

func (r *Repository) GetExpertsByName(title string) ([]model.ArtExpert, error) {
	var experts []model.ArtExpert
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&experts).Error
	if err != nil {
		return nil, err
	}
	return experts, nil
}

func (r *Repository) GetExpertByID(id int) (*model.ArtExpert, error) {
	var expert model.ArtExpert
	err := r.db.First(&expert, id).Error
	if err != nil {
		return nil, err
	}
	return &expert, nil
}

func (r *Repository) ListExperts(title string) ([]model.ArtExpert, error) {
	var experts []model.ArtExpert
	q := r.db
	if title != "" {
		q = q.Where("title ILIKE ?", "%"+title+"%")
	}
	if err := q.Find(&experts).Error; err != nil {
		return nil, err
	}
	return experts, nil
}

func (r *Repository) AddExpert(expert *model.ArtExpert) error {
	return r.db.Create(expert).Error
}

func (r *Repository) UpdateExpert(id uint, title, description, name, algorithm string, status *bool) (*model.ArtExpert, error) {
	var expert model.ArtExpert
	if err := r.db.First(&expert, id).Error; err != nil {
		return nil, err
	}
	if title != "" {
		expert.Title = title
	}
	if description != "" {
		expert.Description = description
	}
	if name != "" {
		expert.Name = name
	}
	if algorithm != "" {
		expert.Algorithm = algorithm
	}
	if status != nil {
		expert.Status = *status
	}
	if err := r.db.Save(&expert).Error; err != nil {
		return nil, err
	}
	return &expert, nil
}

func (r *Repository) DeleteExpert(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var expert model.ArtExpert
		if err := tx.First(&expert, id).Error; err != nil {
			return fmt.Errorf("expert not found: %w", err)
		}

		if expert.ImgURL != nil && *expert.ImgURL != "" {
			oldImageURL, err := url.Parse(*expert.ImgURL)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.minio.bucketName))
				r.minio.client.RemoveObject(context.Background(), r.minio.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		return tx.Delete(&model.ArtExpert{}, id).Error
	})
}

func (r *Repository) SaveExpertImage(ctx context.Context, id uint, fileHeader *multipart.FileHeader) (string, error) {
	var finalURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var expert model.ArtExpert
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&expert, id).Error; err != nil {
			return err
		}

		if expert.ImgURL != nil && *expert.ImgURL != "" {
			oldImageURL, err := url.Parse(*expert.ImgURL)
			if err == nil {
				oldObject := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.minio.bucketName))
				r.minio.client.RemoveObject(ctx, r.minio.bucketName, oldObject, minio.RemoveObjectOptions{})
			}
		}

		// Сформировать имя файла
		fileName := filepath.Base(fileHeader.Filename)
		ext := filepath.Ext(fileName)
		name := fileName[:len(fileName)-len(ext)]
		if name == "" {
			name = "image"
		}
		safeName := strings.ReplaceAll(name, " ", "-")
		fileName = safeName + ext
		objectName := "img/" + fileName

		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = r.minio.client.PutObject(ctx, r.minio.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
		if err != nil {
			return err
		}

		minioEndpoint := os.Getenv("MINIO_ENDPOINT")
		if minioEndpoint == "" {
			minioEndpoint = "127.0.0.1:9000"
		}
		finalURL = fmt.Sprintf("http://%s/%s/%s", minioEndpoint, r.minio.bucketName, objectName)
		if err := tx.Model(&expert).Update("img_url", finalURL).Error; err != nil {
			return err
		}
		return nil
	})
	return finalURL, err
}

func (r *Repository) AddExpertToDraftOrder(expertID, orderID uint) error {
	link := model.ExpertsToOrders{
		ID_artcenter: expertID,
		ID_order:     orderID,
	}
	return r.db.Create(&link).Error
}
