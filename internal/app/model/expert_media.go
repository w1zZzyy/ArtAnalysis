package model

import "time"

// ExpertMedia represents a media file (image/video) attached to an art expert/service
type ExpertMedia struct {
	ID_media     uint      `gorm:"column:id_media;primaryKey;autoIncrement" json:"id_media"`
	ID_artcenter uint      `gorm:"column:id_artcenter;not null" json:"id_artcenter"`
	MediaURL     string    `gorm:"column:media_url;not null" json:"media_url"`
	MediaType    string    `gorm:"column:media_type;not null;default:image" json:"media_type"` // "image" or "video"
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// Связь с ArtExpert
	ArtExpert ArtExpert `gorm:"foreignKey:ID_artcenter;references:ID_artcenter" json:"-"`
}

// TableName указывает имя таблицы для GORM
func (ExpertMedia) TableName() string {
	return "expert_media"
}
