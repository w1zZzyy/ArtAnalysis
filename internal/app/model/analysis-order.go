package model

import "time"

type AnalysisOrder struct {
	ID_order      uint      `gorm:"column:id_order;primaryKey;autoIncrement"`
	ID_creator    uint      `gorm:"column:id_creator;not null"`
	ID_moderator  *uint     `gorm:"column:id_moderator"`
	OrderStatus   string    `gorm:"column:order_status;size:255;not null;default:черновик"`
	DateCreated   time.Time `gorm:"column:date_created;not null"`
	DateFormed    time.Time `gorm:"column:date_formed"`
	DateCompleted time.Time `gorm:"column:date_completed"`

	ResultX *float32 `gorm:"column:result_x"`
	ResultY *float32 `gorm:"column:result_y"`

	ExpertsLinks []ExpertsToOrders `gorm:"foreignKey:ID_order;references:ID_order"`

	Moderator Users `gorm:"foreignKey:ID_moderator;references:ID_user"`
	User      Users `gorm:"foreignKey:ID_creator;references:ID_user"`
}
