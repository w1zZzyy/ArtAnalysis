package model

type Users struct {
	ID_user     uint   `gorm:"column:id_user;primaryKey"`
	Login       string `gorm:"column:login;not null;size:255;unique"`
	Password    string `gorm:"column:password;size:255;not null"`
	IsModerator bool   `gorm:"column:is_moderator;default:false"`

	Orders []AnalysisOrder `gorm:"foreignKey:ID_creator;references:ID_user" json:"-"`
}
