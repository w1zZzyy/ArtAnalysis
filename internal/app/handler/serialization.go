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
	Status      *bool   `json:"status" example:"false"`
	Algorithm   *string `json:"algorithm" example:"geometric_center"`
	ImgURL      *string `json:"img_url" example:"updated_photo.png"`
}

// DTO_Req_OrderUpd запрос обновления заявки
// @Description Обновление полей заявки (например, статус или результаты)
type DTO_Req_OrderUpd struct {
	OrderStatus *string  `json:"order_status" example:"formed" enums:"draft,formed,completed,rejected"`
	ResultX     *float32 `json:"result_x" example:"0.45"`
	ResultY     *float32 `json:"result_y" example:"0.52"`
}

// DTO_Req_OrderResolve запрос модератора для завершения/отклонения заявки
// @Description Завершение или отклонение заявки
type DTO_Req_OrderResolve struct {
	Action string `json:"action" example:"complete" enums:"complete,reject" binding:"required"`
}

// DTO_Req_UpdateExpertInOrder запрос обновления координат эксперта в заявке
// @Description Изменение координат композиционного центра эксперта
type DTO_Req_UpdateExpertInOrder struct {
	CenterX *float32 `json:"center_x" example:"123.45" binding:"required"`
	CenterY *float32 `json:"center_y" example:"234.56" binding:"required"`
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

// DTO_Resp_Order ответ с данными заявки
// @Description Полная информация о заявке с экспертами
type DTO_Resp_Order struct {
	ID_order      uint                   `json:"id_order" example:"5"`
	OrderStatus   string                 `json:"order_status" example:"formed" enums:"draft,formed,completed,rejected"`
	DateCreated   time.Time              `json:"date_created" example:"2025-11-05T12:00:00Z"`
	DateFormed    time.Time              `json:"date_formed" example:"2025-11-06T12:00:00Z"`
	DateCompleted time.Time              `json:"date_completed" example:"2025-11-07T12:00:00Z"`
	ResultX       *float32               `json:"result_x" example:"0.51"`
	ResultY       *float32               `json:"result_y" example:"0.48"`
	ID_creator    uint                   `json:"id_creator" example:"1"`
	ID_moderator  *uint                  `json:"id_moderator" example:"2"`
	Experts       []DTO_Resp_OrderExpert `json:"experts"`
}

// DTO_Resp_OrderExpert связь эксперта и заявки
// @Description Информация об эксперте внутри заявки
type DTO_Resp_OrderExpert struct {
	ID_artcenter uint     `json:"id_artcenter" example:"3"`
	ID_order     uint     `json:"id_order" example:"5"`
	CenterX      *float32 `json:"center_x" example:"123.45"`
	CenterY      *float32 `json:"center_y" example:"234.56"`
	Title        string   `json:"title" example:"Центр композиционного анализа №3"`
	Name         string   `json:"name" example:"Сидоров П.П."`
	Algorithm    string   `json:"algorithm" example:"force_lines"`
}

// DTO_Resp_OrderExpertLink связь задачи и сервиса
// @Description Информация о связи между задачей и гейтом
type DTO_Resp_OrderExpertLink struct {
	OrderID  uint `json:"id_order" example:"1"`
	ExpertID int  `json:"id_artcenter" example:"3"`
}

// DTO_Resp_CurrTaskInfo информация о текущей задаче
// @Description Статистика по текущей задаче пользователя
type DTO_Resp_CurrTaskInfo struct {
	OrderID      uint `json:"order_id" example:"5"`
	ExpertsCount int  `json:"experts_count" example:"3"`
}

// DTO_Resp_CurrentDraftOrder информация о текущем черновике
// @Description Черновая заявка пользователя и количество экспертов
type DTO_Resp_CurrentDraftOrder struct {
	OrderID      uint `json:"order_id" example:"5"`
	ExpertsCount int  `json:"experts_count" example:"2"`
}

// DTO_Resp_SimpleID простой ответ с ID
// @Description Ответ, содержащий только идентификатор
type DTO_Resp_SimpleID struct {
	ID int `json:"id" example:"1"`
}

// DTO_Req_OrderUpdate хранит данные для обновления заявки
type DTO_Req_OrderUpdate struct {
	Status      *string `json:"status"`
	ModeratorID *uint   `json:"moderator_id"`
}

// DTO_Resp_UploadImg результат загрузки изображения эксперта
// @Description Результат загрузки изображения эксперта
type DTO_Resp_UploadImg struct {
	ID    uint   `json:"id" example:"1"`
	Image string `json:"image" example:"expert_image.png"`
}

// DTO_Resp_UpdateExpert ответ обновления градусов
// @Description Результат обновления углов поворота гейта
type DTO_Resp_UpdateExpert struct {
	OrderID  int      `json:"order_id" example:"1"`
	ExpertID int      `json:"expert_id" example:"2"`
	CenterX  *float32 `json:"center_x" example:"45.5"`
	CenterY  *float32 `json:"center_y" example:"45.5"`
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
