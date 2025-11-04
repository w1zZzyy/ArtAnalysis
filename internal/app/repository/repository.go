package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
	// minio client is optional; initialized when env is present
	minio *minioClient
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	r := &Repository{db: db}
	// try to init minio from env vars, ignore errors to keep app running without object storage
	if mc, err := newMinioClientFromEnv(); err == nil {
		r.minio = mc
	}
	return r, nil
}
