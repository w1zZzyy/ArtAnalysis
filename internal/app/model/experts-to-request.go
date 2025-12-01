package model

type ExpertsToRequest struct {
	ID_artcenter uint `gorm:"primaryKey;column:id_artcenter;not null"`
	ID_request   uint `gorm:"primaryKey;column:id_request;not null"`

	CenterX *float32
	CenterY *float32

	ArtExpert     ArtExpert     `gorm:"foreignKey:ID_artcenter;references:ID_artcenter"`
	CenterRequest CenterRequest `gorm:"foreignKey:ID_request;references:ID_request"`
}
