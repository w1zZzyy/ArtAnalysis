package handler

import "time"

//====== REQUESTS ======

// DTO_Req_UserReg запрос регистрации пользователя
// @Description Данные для регистрации нового пользователя
type DTO_Req_UserReg struct {
	Login    string `json:"login" example:"art_user" binding:"required"`
	Password string `json:"password" example:"secure_password_123" binding:"required"`
}

// DTO_Req_UserUpd запрос обновления пользователя
// @Description Данные для обновления пароля пользователя
type DTO_Req_UserUpd struct {
	Password *string `json:"password" example:"new_secure_password"`
}

// DTO_Req_ExpertCreate запрос создания эксперта
// @Description Данные для добавления нового эксперта
type DTO_Req_ExpertCreate struct {
	Title       string  `json:"title" example:"Центр композиционного анализа №1" binding:"required"`
	Description string  `json:"description" example:"Использует метод яркостного центра"`
	Name        string  `json:"name" example:"Иванов И.И." binding:"required"`
	Algorithm   string  `json:"algorithm" example:"luminosity_center" binding:"required"`
	Status      *bool   `json:"status" example:"true"`
	ImgURL      *string `json:"img_url" example:"expert_photo.png"`
}

// DTO_Req_ExpertUpd запрос обновления данных эксперта
// @Description Обновление описания, статуса или изображения эксперта
type DTO_Req_ExpertUpd struct {
	Description *string `json:"description" example:"Измененное описание эксперта"`
	Algorithm   *string `json:"algorithm" example:"geometric_center"`
	ImgURL      *string `json:"img_url" example:"updated_photo.png"`
}

// DTO_Req_CenterRequestUpd запрос обновления заявки
// @Description Обновление полей заявки (например, статус или результаты)
type DTO_Req_CenterRequestUpd struct {
	Description string   `json:"request_description" example:"Calculate center" binding:"required"`
	CenterX     *float32 `json:"center_x" example:"123.45" binding:"required"`
	CenterY     *float32 `json:"center_y" example:"234.56" binding:"required"`
}

// DTO_Req_CenterRequestResolve запрос модератора для завершения/отклонения заявки
// @Description Завершение или отклонение заявки
type DTO_Req_CenterRequestResolve struct {
	Action string `json:"action" example:"complete" enums:"complete,reject" binding:"required"`
}

// DTO_Req_UpdateExpertInRequest запрос обновления координат эксперта в заявке
// @Description Изменение координат композиционного центра эксперта
type DTO_Req_UpdateExpertInRequest struct {
	CenterX *float32 `json:"center_x" example:"123.45" binding:"required"`
	CenterY *float32 `json:"center_y" example:"234.56" binding:"required"`
}

// DTO_Req_AnalysisResult запрос с результатом от async сервиса
// @Description Результат асинхронного анализа от внешнего сервиса
type DTO_Req_AnalysisResult struct {
	RequestID       int      `json:"request_id" binding:"required"`
	Success         bool     `json:"success"`
	AnalysisResult  *string  `json:"analysis_result"`
	ConfidenceScore *float32 `json:"confidence_score"`
	ProcessingTime  float32  `json:"processing_time"`
	Message         string   `json:"message"`
	ServiceKey      string   `json:"service_key" binding:"required"`
}

//====== RESPONSES ======

// DTO_Resp_Expert ответ с данными эксперта
// @Description Полная информация об эксперте
type DTO_Resp_Expert struct {
	ID_artcenter uint    `json:"id_artcenter" example:"1"`
	Title        string  `json:"title" example:"Центр анализа №1"`
	Description  string  `json:"description" example:"Определяет композиционный центр"`
	Status       bool    `json:"status" example:"true"`
	Name         string  `json:"name" example:"Иванов И.И."`
	Algorithm    string  `json:"algorithm" example:"geometric_center"`
	ImgURL       *string `json:"img_url" example:"expert_photo.png"`
}

// DTO_Resp_UploadImg ответ загрузки изображения
// @Description Результат загрузки изображения для эксперта
type DTO_Resp_UploadImg struct {
	ID    int    `json:"id" example:"1"`
	Image string `json:"image" example:"expert_image.png"`
}

// DTO_Resp_CurrCenterRequestInfo информация о текущей задаче
// @Description Статистика по текущей задаче пользователя
type DTO_Resp_CurrCenterRequestInfo struct {
	RequestID    uint `json:"id_request" example:"5"`
	ExpertsCount int  `json:"experts_count" example:"3"`
}

// DTO_Resp_Request ответ с данными заявки
// @Description Полная информация о заявке с экспертами
type DTO_Resp_CenterRequest struct {
	ID_request      uint                           `json:"id_request" example:"5"`
	ID_user         uint                           `json:"id_user" example:"1"`
	RequestStatus   string                         `json:"request_status" example:"formed" enums:"draft,formed,completed,rejected"`
	DateCreated     time.Time                      `json:"date_created" example:"2025-11-05T12:00:00Z"`
	DateFormed      *time.Time                     `json:"date_formed,omitempty" example:"2025-11-06T12:00:00Z"`
	DateConclusion  *time.Time                     `json:"date_conclusion,omitempty" example:"2025-11-07T12:00:00Z"`
	Description     string                         `gorm:"column:description" json:"description" example:"Art state calculation"`
	FactorX         *float32                       `json:"factor_x" example:"0.51"`
	FactorY         *float32                       `json:"factor_y" example:"0.48"`
	AnalysisResult  *string                        `json:"analysis_result,omitempty" example:"Отлично сбалансированная композиция"`
	ConfidenceScore *float32                       `json:"confidence_score,omitempty" example:"0.85"`
	AnalysisSuccess *bool                          `json:"analysis_success,omitempty" example:"true"`
	Experts         []DTO_Resp_CenterRequestExpert `json:"experts"`
}

// DTO_Resp_CenterRequestExpert связь эксперта и заявки
// @Description Информация об эксперте внутри заявки
type DTO_Resp_CenterRequestExpert struct {
	ID_artcenter uint     `json:"id_artcenter" example:"3"`
	ID_request   uint     `json:"id_request" example:"5"`
	CenterX      *float32 `json:"center_x" example:"123.45"`
	CenterY      *float32 `json:"center_y" example:"234.56"`
	Title        string   `json:"title" example:"Центр композиционного анализа №3"`
	Name         string   `json:"name" example:"Сидоров П.П."`
	Algorithm    string   `json:"algorithm" example:"force_lines"`
}

// DTO_Resp_CenterRequestExpertLink связь задачи и сервиса
// @Description Информация о связи между задачей и эксперта
type DTO_Resp_CenterRequestExpertLink struct {
	RequestID uint `json:"id_request" example:"1"`
	ExpertID  int  `json:"id_artcenter" example:"3"`
}

// DTO_Resp_CurrentDraftCenterRequest информация о текущем черновике
// @Description Черновая заявка пользователя и количество экспертов
type DTO_Resp_CurrentDraftCenterRequest struct {
	RequestID    uint `json:"id_request" example:"5"`
	ExpertsCount int  `json:"experts_count" example:"2"`
}

// DTO_Resp_SimpleID простой ответ с ID
// @Description Ответ, содержащий только идентификатор
type DTO_Resp_SimpleID struct {
	ID int `json:"id" example:"1"`
}

// DTO_Resp_Update ответ обновления центра
// @Description Результат обновления центра
type DTO_Resp_Update struct {
	RequestID int      `json:"request_id" example:"1"`
	ExpertID  int      `json:"expert_id" example:"2"`
	CenterX   *float32 `json:"center_x" example:"45.5"`
	CenterY   *float32 `json:"center_y" example:"45.5"`
}

// DTO_User базовая информация о пользователе
// @Description Информация о пользователе системы
type DTO_User struct {
	ID_user uint   `json:"id_user" example:"1"`
	Login   string `json:"login" example:"art_researcher"`
	IsAdmin bool   `json:"is_admin" example:"false"`
}

// DTO_Resp_User ответ с пользователем
// @Description Упрощенные данные пользователя
type DTO_Resp_User struct {
	Login string `json:"login" example:"art_researcher"`
}

// DTO_Resp_TokenLogin ответ аутентификации
// @Description JWT токен и данные пользователя
type DTO_Resp_TokenLogin struct {
	Token string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  DTO_User `json:"user"`
}

// DTO_Resp_ExpertMedia медиафайл эксперта
// @Description Информация о медиафайле (фото/видео) эксперта
type DTO_Resp_ExpertMedia struct {
	ID_media     uint   `json:"id_media" example:"1"`
	ID_artcenter uint   `json:"id_artcenter" example:"1"`
	MediaURL     string `json:"media_url" example:"http://localhost:9000/art-center/media/photo.jpg"`
	MediaType    string `json:"media_type" example:"image" enums:"image,video"`
	CreatedAt    string `json:"created_at" example:"2025-11-05T12:00:00Z"`
}

// DTO_Resp_ExpertWithMedia эксперт с медиафайлами
// @Description Полная информация об эксперте включая медиафайлы
type DTO_Resp_ExpertWithMedia struct {
	ID_artcenter uint                   `json:"id_artcenter" example:"1"`
	Title        string                 `json:"title" example:"Центр анализа №1"`
	Description  string                 `json:"description" example:"Определяет композиционный центр"`
	Status       bool                   `json:"status" example:"true"`
	Name         string                 `json:"name" example:"Иванов И.И."`
	Algorithm    string                 `json:"algorithm" example:"geometric_center"`
	ImgURL       *string                `json:"img_url" example:"expert_photo.png"`
	Media        []DTO_Resp_ExpertMedia `json:"media"`
}
