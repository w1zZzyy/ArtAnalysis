package model

import "time"

type CenterRequest struct {
	ID_request         uint       `gorm:"column:id_request;primaryKey;autoIncrement"`
	ID_creator         uint       `gorm:"column:id_creator;not null"`
	ID_moderator       *uint      `gorm:"column:id_moderator"`
	RequestStatus      string     `gorm:"column:request_status;size:255;not null;default:черновик"`
	DateCreated        time.Time  `gorm:"column:date_created;not null"`
	DateFormed         *time.Time `gorm:"column:date_formed"`
	DateConclusion     *time.Time `gorm:"column:date_conclusion"`
	RequestDescription string     `gorm:"column:request_description"`

	FactorX *float32 `gorm:"column:factor_x"`
	FactorY *float32 `gorm:"column:factor_y"`

	// Поле для результата асинхронного анализа (заполняется async сервисом)
	AnalysisResult  *string  `gorm:"column:analysis_result"`
	ConfidenceScore *float32 `gorm:"column:confidence_score"`
	AnalysisSuccess *bool    `gorm:"column:analysis_success"`

	ExpertsLinks []ExpertsToRequest `gorm:"foreignKey:ID_request;references:ID_request"`

	Moderator Users `gorm:"foreignKey:ID_moderator;references:ID_user"`
	User      Users `gorm:"foreignKey:ID_creator;references:ID_user"`
}
