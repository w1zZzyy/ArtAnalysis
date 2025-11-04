package model

type ArtExpert struct {
	ID_artcenter uint   `gorm:"column:id_artcenter;primaryKey;autoIncrement"`
	Title        string `gorm:"column:title;size:255;not null;default:artcenter-no-name;unique"`
	Description  string `gorm:"column:description;not null"`
	Status       bool   `gorm:"column:status;not null;default:true"`
	ImgURL       *string
	Name         string `gorm:"column:name;not null"`
	Algorithm    string `gorm:"column:algorithm;not null"`

	Orders []ExpertsToOrders `gorm:"foreignKey:ID_artcenter;references:ID_artcenter"`
}
