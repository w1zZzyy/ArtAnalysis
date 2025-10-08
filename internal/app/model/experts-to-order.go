package model

type ExpertsToOrders struct {
	ID_artcenter uint `gorm:"primaryKey;column:id_artcenter;not null"`
	ID_order     uint `gorm:"primaryKey;column:id_order;not null"`

	CenterX *float32
	CenterY *float32

	ArtExpert     ArtExpert     `gorm:"foreignKey:ID_artcenter;references:ID_artcenter"`
	AnalysisOrder AnalysisOrder `gorm:"foreignKey:ID_order;references:ID_order"`
}
