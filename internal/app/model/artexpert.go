package model

type ArtExpert struct {
	ID_artcenter uint   `gorm:"column:id_artcenter;primaryKey"`
	Title        string `gorm:"column:title;size:255;not null;default:artcenter-no-name;unique"`
	Description  string `gorm:"column:description;not null"`
	Status       bool   `gorm:"column:status;not null;default:true"`
	ImgURL       *string
	Algorithm    string `gorm:"column:algorithm;not null"`

	Orders []ExpertsToOrders `gorm:"foreignKey:ID_artcenter;references:ID_artcenter"`
}
