package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Float32Ptr(f float32) *float32 {
	return &f
}

func (h *Handler) AddExpertToRequest(c *gin.Context) {
	expertID, err := strconv.Atoi(c.Param("id_expert"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	request, err := h.Repository.GetDraftRequest(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newReq := model.CenterRequest{
			ID_creator:    userID,
			RequestStatus: model.StatusDraft,
			DateCreated:   time.Now(),
			FactorX:       Float32Ptr(1.0),
			FactorY:       Float32Ptr(1.0),
		}
		if createErr := h.Repository.CreateRequest(&newReq); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		request = &newReq
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	if err = h.Repository.AddExpertToRequest(request.ID_request, uint(expertID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/ArtAnalysis")
}

func (h *Handler) GetCenterRequest(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.GetRequestWithExperts(uint(requestID))
	if err != nil {
		c.HTML(http.StatusNotFound, "invalid_page.html", nil)
		return
	}

	if len(request.ExpertsLinks) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty request, add experts first"))
		return
	}

	c.HTML(http.StatusOK, "request.html", request)
}

// DeleteCenterRequest выполняет логическое удаление заявки
func (h *Handler) DeleteCenterRequest(c *gin.Context) {
	requestID, _ := strconv.Atoi(c.Param("id"))

	if err := h.Repository.LogicallyDeleteRequest(uint(requestID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/ArtAnalysis")
}

// ---- JSON API for tasks ----

// ApiListCenterRequest godoc
// @Summary Получить список заявок на анализ
// @Description Возвращает список заявок. Пользователь видит только свои заявки, модератор — все заявки. Неавторизованный пользователь получает ошибку доступа.
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param from query string false "Начальная дата (фильтр от)"
// @Param to query string false "Конечная дата (фильтр до)"
// @Success 200 {array} DTO_Resp_CenterRequest "Список заявок"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal server error"
// @Security BearerAuth
// @Router /api/center_request [get]
func (h *Handler) ApiListCenterRequest(ctx *gin.Context) {
	// Получаем ID пользователя из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("требуется авторизация"))
		return
	}

	// Получаем информацию о пользователе
	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	status := ctx.Query("status")
	from := ctx.Query("from")
	to := ctx.Query("to")

	var reqs []*model.CenterRequest

	if user.IsModerator {
		reqs, err = h.Repository.ListCenterRequests(status, from, to)
	} else {
		reqs, err = h.Repository.ListRequestsByUser(userID, status, from, to)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Преобразуем заявки в DTO
	var representRequests []DTO_Resp_CenterRequest
	for _, req := range reqs {
		var dtoExperts []DTO_Resp_CenterRequestExpert
		for _, link := range req.ExpertsLinks {
			dtoExperts = append(dtoExperts, DTO_Resp_CenterRequestExpert{
				ID_artcenter: link.ID_artcenter,
				ID_request:   link.ID_request,
				CenterX:      link.CenterX,
				CenterY:      link.CenterY,
				Title:        link.ArtExpert.Title,
				Name:         link.ArtExpert.Name,
				Algorithm:    link.ArtExpert.Algorithm,
			})
		}

		representRequests = append(representRequests, DTO_Resp_CenterRequest{
			ID_request:     req.ID_request,
			ID_user:        req.ID_creator,
			RequestStatus:  req.RequestStatus,
			DateCreated:    req.DateCreated,
			DateFormed:     req.DateFormed,
			DateConclusion: req.DateConclusion,
			FactorX:        req.FactorX,
			FactorY:        req.FactorY,
			Experts:        dtoExperts,
		})
	}

	ctx.JSON(http.StatusOK, representRequests)
}

// ApiGetCenterRequestByID возвращает заявку по ID
// @Summary Получить заявку по ID
// @Description Возвращает детальную информацию о заявке на анализ по её идентификатору
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} DTO_Resp_CenterRequest "Детали заявки"
// @Failure 400 {object} string "Invalid request ID"
// @Failure 404 {object} string "request not found"
// @Router /api/center_request/current [get]
func (h *Handler) ApiGetCenterRequestByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	req, err := h.Repository.GetRequestWithExperts(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	var represent_req DTO_Resp_CenterRequest
	var dtoCenterRequestExpert []DTO_Resp_CenterRequestExpert
	for _, expert := range req.ExpertsLinks {
		dtoCenterRequestExpert = append(dtoCenterRequestExpert, DTO_Resp_CenterRequestExpert{
			ID_artcenter: expert.ID_artcenter,
			ID_request:   expert.ID_request,
			CenterX:      expert.CenterX,
			CenterY:      expert.CenterY,
		})
	}

	represent_req = DTO_Resp_CenterRequest{
		ID_request:     req.ID_request,
		RequestStatus:  req.RequestStatus,
		DateCreated:    req.DateCreated,
		ID_user:        req.ID_creator,
		DateConclusion: req.DateConclusion,
		FactorX:        req.FactorX,
		FactorY:        req.FactorY,
		Experts:        dtoCenterRequestExpert,
	}

	ctx.JSON(http.StatusOK, represent_req)
}

// ApiUpdateCenterRequest обновляет данные заявки
// @Summary Обновить заявку
// @Description Обновляет статус заявки и/или модератора, если это требуется
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body DTO_Req_CenterRequest true "Данные для обновления"
// @Success 200 {object} DTO_Resp_CenterRequest "Обновленная заявка"
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal server error"
// @Router /api/center_request/{id} [put]
func (h *Handler) ApiUpdateCenterRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req DTO_Req_CenterRequestUpd
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	center_request, err := h.Repository.UpdateCenterRequest(uint(id), req.Description)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var represent_req DTO_Resp_CenterRequest
	var dtoCenterRequestExpert []DTO_Resp_CenterRequestExpert
	for _, expert := range center_request.ExpertsLinks {
		dtoCenterRequestExpert = append(dtoCenterRequestExpert, DTO_Resp_CenterRequestExpert{
			ID_artcenter: expert.ID_artcenter,
			ID_request:   expert.ID_request,
			CenterX:      expert.CenterX,
			CenterY:      expert.CenterY,
		})
	}

	represent_req = DTO_Resp_CenterRequest{
		ID_request:     center_request.ID_request,
		RequestStatus:  center_request.RequestStatus,
		DateCreated:    center_request.DateCreated,
		ID_user:        center_request.ID_creator,
		DateConclusion: center_request.DateConclusion,
		FactorX:        center_request.FactorX,
		FactorY:        center_request.FactorY,
		Experts:        dtoCenterRequestExpert,
	}

	ctx.JSON(http.StatusOK, represent_req)
}

// ApiFormCenterRequest формирует заявку
// @Summary Сформировать заявку
// @Description Переводит заявку из статуса черновика в статус сформированной
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} DTO_Resp_CenterRequest "Сформированная заявка"
// @Failure 400 {object} string "Invalid request ID or cannot form request"
// @Router /api/center_request/{id}/form [put]
func (h *Handler) ApiFormCenterRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.FormCenterRequest(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var dtoExperts []DTO_Resp_CenterRequestExpert
	for _, link := range request.ExpertsLinks {
		dtoExperts = append(dtoExperts, DTO_Resp_CenterRequestExpert{
			ID_artcenter: link.ID_artcenter,
			ID_request:   link.ID_request,
			CenterX:      link.CenterX,
			CenterY:      link.CenterY,
			Title:        link.ArtExpert.Title,
			Name:         link.ArtExpert.Name,
			Algorithm:    link.ArtExpert.Algorithm,
		})
	}

	dtoRequest := DTO_Resp_CenterRequest{
		ID_request:     request.ID_request,
		RequestStatus:  request.RequestStatus,
		DateCreated:    request.DateCreated,
		DateFormed:     request.DateFormed,
		DateConclusion: request.DateConclusion,
		FactorX:        request.FactorX,
		FactorY:        request.FactorY,
		ID_user:        request.ID_creator,
		Experts:        dtoExperts,
	}

	ctx.JSON(http.StatusOK, dtoRequest)
}

// ApiResolveCenterRequest завершает или отклоняет заявку
// @Summary Завершить/отклонить заявку
// @Description Выполняет завершение или отклонение заявки по анализу произведения искусства. При завершении рассчитывается финальный композиционный центр.
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body object true "Действие с заявкой (action: complete | reject)"
// @Success 200 {object} DTO_Resp_CenterRequest "Обновлённая заявка"
// @Failure 400 {object} string "Invalid input or missing required fields"
// @Failure 401 {object} string "Unauthorized"
// @Failure 404 {object} string "Request not found"
// @Failure 500 {object} string "Internal server error"
// @Router /api/center_request/{id}/resolve [put]
func (h *Handler) ApiResolveCenterRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid request ID: %v", err))
		return
	}

	var req DTO_Req_CenterRequestResolve

	if err := ctx.ShouldBindJSON(&req); err != nil {
		if err.Error() == "EOF" {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("request body is empty, expected JSON with 'action' field"))
		} else {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err))
		}
		return
	}

	if req.Action == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("action field is required"))
		return
	}

	// Проверяем, что action имеет допустимое значение
	if req.Action != "complete" && req.Action != "reject" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("action must be 'complete' or 'reject'"))
		return
	}

	// Если задача завершается, вычисляем результат
	if req.Action == "complete" {
		// Проверяем, что у задачи есть все необходимые параметры для вычисления результата
		req, err := h.Repository.GetRequestWithExperts(uint(id))
		if err != nil {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}

		// Проверяем наличие описания задачи
		if req.RequestDescription == "" || req.FactorX == nil || req.FactorY == nil {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("task description is required for completion"))
			return
		}

		err = h.Repository.CalculateArtAnalysis(uint(id))
		if err != nil {
			logrus.Errorf("Failed to calculate result for task %d: %v", id, err)
			h.errorHandler(ctx, http.StatusInternalServerError,
				fmt.Errorf("failed to calculate task result: %v", err))
			return
		}
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	adminID := uint(userID)
	center_request, err := h.Repository.ResolveCenterRequest(uint(id), adminID, req.Action, time.Now())
	if err != nil {
		logrus.Errorf("Failed to resolve task %d: %v", id, err)
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var represent_req DTO_Resp_CenterRequest
	var dtoCenterRequestExpert []DTO_Resp_CenterRequestExpert
	for _, expert := range center_request.ExpertsLinks {
		dtoCenterRequestExpert = append(dtoCenterRequestExpert, DTO_Resp_CenterRequestExpert{
			ID_artcenter: expert.ID_artcenter,
			ID_request:   expert.ID_request,
			CenterX:      expert.CenterX,
			CenterY:      expert.CenterY,
		})
	}

	represent_req = DTO_Resp_CenterRequest{
		ID_request:     center_request.ID_request,
		RequestStatus:  center_request.RequestStatus,
		DateCreated:    center_request.DateCreated,
		ID_user:        center_request.ID_creator,
		DateConclusion: center_request.DateConclusion,
		FactorX:        center_request.FactorX,
		FactorY:        center_request.FactorY,
		Experts:        dtoCenterRequestExpert,
	}

	ctx.JSON(http.StatusOK, represent_req)
}

// ApiDeleteCenterRequest удаляет заказ на анализ
// @Summary Удалить заказ на анализ
// @Description Полностью удаляет заказ анализа произведения искусства
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {object} DTO_Resp_SimpleID "ID удалённого заказа"
// @Failure 400 {object} string "Invalid request ID"
// @Failure 500 {object} string "Internal server error"
// @Router /api/center_request/{id} [delete]
func (h *Handler) ApiDeleteCenterRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteCenterRequest(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_SimpleID{ID: id})
}

// ApiRemoveExpertFromCenterRequest удаляет эксперта из заказа
// @Summary Удалить эксперта из заказа
// @Description Удаляет связь между экспертом и заказом анализа
// @Tags M-M
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Param expert_id path int true "ID эксперта"
// @Success 200 {object} DTO_Resp_CenterRequestExpertLink "Информация об удалённой связи"
// @Failure 400 {object} string "Invalid IDs"
// @Failure 500 {object} string "Internal server error"
// @Router /api/center_request/{id}/experts/{expert_id} [delete]
func (h *Handler) ApiRemoveExpertFromCenterRequest(ctx *gin.Context) {
	requestID, err1 := strconv.Atoi(ctx.Param("id"))
	expertID, err2 := strconv.Atoi(ctx.Param("expert_id"))
	if err1 != nil || err2 != nil || requestID <= 0 || expertID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid ids"))
		return
	}

	if err := h.Repository.RemoveExpertFromRequest(uint(requestID), uint(expertID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_CenterRequestExpertLink{
		RequestID: uint(requestID),
		ExpertID:  int(expertID),
	})
}

// ApiUpdateExpertInCenterRequest обновляет данные эксперта в заказе
// @Summary Обновить параметры эксперта
// @Description Обновляет координаты центра, переданные экспертом, в рамках заказа анализа
// @Tags M-M
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Param expert_id path int true "ID эксперта"
// @Param request DTO_Req_CenterRequestUpd true "Новые координаты эксперта"
// @Success 200 {object} DTO_Resp_Update "Обновлённые данные эксперта"
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal server error"
// @Router /api/center_request/{id}/experts/{expert_id} [put]
func (h *Handler) ApiUpdateExpertInCenterRequest(ctx *gin.Context) {
	taskID, err1 := strconv.Atoi(ctx.Param("request_id"))
	expertID, err2 := strconv.Atoi(ctx.Param("expert_id"))
	if err1 != nil || err2 != nil || taskID <= 0 || expertID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid ids"))
		return
	}

	var req DTO_Req_CenterRequestUpd
	if err := ctx.ShouldBindJSON(&req); err != nil || (req.CenterX == nil && req.CenterY == nil) {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("center_x or center_y required"))
		return
	}

	if err := h.Repository.UpdateExpertInRequest(uint(taskID), uint(expertID), req.CenterX, req.CenterY); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_Update{
		RequestID: taskID,
		ExpertID:  expertID,
		CenterX:   req.CenterX,
		CenterY:   req.CenterY,
	})
}
