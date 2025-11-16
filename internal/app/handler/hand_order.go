package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hardcodedUserID = 1

// AddExpertToOrder добавляет эксперта в черновик заявки
func (h *Handler) AddExpertToOrder(c *gin.Context) {
	expertID, err := strconv.Atoi(c.Param("id_expert"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	// Получаем черновик заявки пользователя
	order, err := h.Repository.GetDraftOrder(hardcodedUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Если черновика нет — создаём новый
		newOrder := model.AnalysisOrder{
			ID_creator:  hardcodedUserID,
			OrderStatus: model.StatusDraft,
		}
		if createErr := h.Repository.CreateOrder(&newOrder); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		order = &newOrder
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	// Добавляем эксперта в заявку
	if err = h.Repository.AddExpertToOrder(order.ID_order, uint(expertID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/experts")
}

// GetOrder отображает заявку с экспертами
func (h *Handler) GetOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	order, err := h.Repository.GetOrderWithExperts(uint(orderID))
	if err != nil {
		c.HTML(http.StatusOK, "analysis.html", gin.H{
			"ID_order":     nil, // можно показать ID даже если нет данных
			"ExpertsLinks": nil,
			"ResultX":      nil,
			"ResultY":      nil,
		})
		return
	}

	if len(order.ExpertsLinks) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty order, add experts first"))
		return
	}

	c.HTML(http.StatusOK, "analysis.html", order)
}

// DeleteOrder выполняет логическое удаление заявки
func (h *Handler) DeleteOrder(c *gin.Context) {
	orderID, _ := strconv.Atoi(c.Param("order_id"))

	if err := h.Repository.LogicallyDeleteOrder(uint(orderID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/experts")
}

func (h *Handler) ApiGetCurrentDraftOrder(ctx *gin.Context) {
	order, count, err := h.Repository.GetCurrentDraftOrder(hardcodedUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, gin.H{"draft_order": order, "experts_count": count})
}

// ApiListAnalysisOrders godoc
// @Summary Получить список заявок на анализ
// @Description Возвращает список заявок. Пользователь видит только свои заявки, модератор — все заявки. Неавторизованный пользователь получает ошибку доступа.
// @Tags AnalysisOrders
// @Accept json
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param from query string false "Начальная дата (фильтр от)"
// @Param to query string false "Конечная дата (фильтр до)"
// @Success 200 {array} DTO_Resp_Order "Список заявок"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal server error"
// @Security BearerAuth
// @Router /api/analysis_orders [get]
func (h *Handler) ApiListAnalysisOrders(ctx *gin.Context) {
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

	var orders []*model.AnalysisOrder

	// Разделяем логику в зависимости от роли пользователя
	if user.IsModerator {
		// Модератор видит все заявки
		orders, err = h.Repository.ListAnalysisOrders(status, from, to)
	} else {
		// Обычный пользователь видит только свои заявки
		orders, err = h.Repository.ListTaskByUser(userID, status, from, to)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Преобразуем заявки в DTO
	var representOrders []DTO_Resp_Order
	for _, order := range orders {
		var dtoExperts []DTO_Resp_OrderExpert
		for _, link := range order.ExpertsLinks {
			dtoExperts = append(dtoExperts, DTO_Resp_OrderExpert{
				ID_artcenter: link.ID_artcenter,
				ID_order:     link.ID_order,
				CenterX:      link.CenterX,
				CenterY:      link.CenterY,
				Title:        link.ArtExpert.Title,
				Name:         link.ArtExpert.Name,
				Algorithm:    link.ArtExpert.Algorithm,
			})
		}

		representOrders = append(representOrders, DTO_Resp_Order{
			ID_order:      order.ID_order,
			OrderStatus:   order.OrderStatus,
			DateCreated:   order.DateCreated,
			DateFormed:    order.DateFormed,
			DateCompleted: order.DateCompleted,
			ResultX:       order.ResultX,
			ResultY:       order.ResultY,
			ID_creator:    order.ID_creator,
			ID_moderator:  order.ID_moderator,
			Experts:       dtoExperts,
		})
	}

	ctx.JSON(http.StatusOK, representOrders)
}

// ApiGetAnalysisOrderByID возвращает заявку по ID
// @Summary Получить заявку по ID
// @Description Возвращает детальную информацию о заявке на анализ по её идентификатору
// @Tags AnalysisOrders
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} DTO_Resp_Order "Детали заявки"
// @Failure 400 {object} string "Invalid order ID"
// @Failure 404 {object} string "Order not found"
// @Router /api/analysis_orders/{id} [get]
func (h *Handler) ApiGetAnalysisOrderByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	order, err := h.Repository.GetAnalysisOrderByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Преобразуем экспертов в DTO
	var dtoExperts []DTO_Resp_OrderExpert
	for _, link := range order.ExpertsLinks {
		dtoExperts = append(dtoExperts, DTO_Resp_OrderExpert{
			ID_artcenter: link.ID_artcenter,
			ID_order:     link.ID_order,
			CenterX:      link.CenterX,
			CenterY:      link.CenterY,
			Title:        link.ArtExpert.Title,
			Name:         link.ArtExpert.Name,
			Algorithm:    link.ArtExpert.Algorithm,
		})
	}

	// Собираем DTO заявки
	dtoOrder := DTO_Resp_Order{
		ID_order:      order.ID_order,
		OrderStatus:   order.OrderStatus,
		DateCreated:   order.DateCreated,
		DateFormed:    order.DateFormed,
		DateCompleted: order.DateCompleted,
		ResultX:       order.ResultX,
		ResultY:       order.ResultY,
		ID_creator:    order.ID_creator,
		ID_moderator:  order.ID_moderator,
		Experts:       dtoExperts,
	}

	ctx.JSON(http.StatusOK, dtoOrder)
}

// ApiUpdateAnalysisOrder обновляет данные заявки
// @Summary Обновить заявку
// @Description Обновляет статус заявки и/или модератора, если это требуется
// @Tags AnalysisOrders
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body DTO_Req_OrderUpdate true "Данные для обновления"
// @Success 200 {object} DTO_Resp_Order "Обновленная заявка"
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal server error"
// @Router /api/analysis_orders/{id} [put]
func (h *Handler) ApiUpdateAnalysisOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var body struct {
		Status      *string `json:"status"`
		ModeratorID *uint   `json:"moderator_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	order, err := h.Repository.UpdateAnalysisOrder(uint(id), body.Status, body.ModeratorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Преобразуем экспертов в DTO
	var dtoExperts []DTO_Resp_OrderExpert
	for _, link := range order.ExpertsLinks {
		dtoExperts = append(dtoExperts, DTO_Resp_OrderExpert{
			ID_artcenter: link.ID_artcenter,
			ID_order:     link.ID_order,
			CenterX:      link.CenterX,
			CenterY:      link.CenterY,
			Title:        link.ArtExpert.Title,
			Name:         link.ArtExpert.Name,
			Algorithm:    link.ArtExpert.Algorithm,
		})
	}

	dtoOrder := DTO_Resp_Order{
		ID_order:      order.ID_order,
		OrderStatus:   order.OrderStatus,
		DateCreated:   order.DateCreated,
		DateFormed:    order.DateFormed,
		DateCompleted: order.DateCompleted,
		ResultX:       order.ResultX,
		ResultY:       order.ResultY,
		ID_creator:    order.ID_creator,
		ID_moderator:  order.ID_moderator,
		Experts:       dtoExperts,
	}

	ctx.JSON(http.StatusOK, dtoOrder)
}

// ApiFormAnalysisOrder формирует заявку
// @Summary Сформировать заявку
// @Description Переводит заявку из статуса черновика в статус сформированной
// @Tags AnalysisOrders
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} DTO_Resp_Order "Сформированная заявка"
// @Failure 400 {object} string "Invalid order ID or cannot form order"
// @Router /api/analysis_orders/{id}/form [put]
func (h *Handler) ApiFormAnalysisOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	order, err := h.Repository.FormAnalysisOrder(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Преобразуем экспертов в DTO
	var dtoExperts []DTO_Resp_OrderExpert
	for _, link := range order.ExpertsLinks {
		dtoExperts = append(dtoExperts, DTO_Resp_OrderExpert{
			ID_artcenter: link.ID_artcenter,
			ID_order:     link.ID_order,
			CenterX:      link.CenterX,
			CenterY:      link.CenterY,
			Title:        link.ArtExpert.Title,
			Name:         link.ArtExpert.Name,
			Algorithm:    link.ArtExpert.Algorithm,
		})
	}

	dtoOrder := DTO_Resp_Order{
		ID_order:      order.ID_order,
		OrderStatus:   order.OrderStatus,
		DateCreated:   order.DateCreated,
		DateFormed:    order.DateFormed,
		DateCompleted: order.DateCompleted,
		ResultX:       order.ResultX,
		ResultY:       order.ResultY,
		ID_creator:    order.ID_creator,
		ID_moderator:  order.ID_moderator,
		Experts:       dtoExperts,
	}

	ctx.JSON(http.StatusOK, dtoOrder)
}

// ApiResolveAnalysisOrder завершает или отклоняет заявку
// @Summary Завершить/отклонить заявку
// @Description Выполняет завершение или отклонение заявки по анализу произведения искусства. При завершении рассчитывается финальный композиционный центр.
// @Tags AnalysisOrders
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body object true "Действие с заявкой (action: complete | reject)"
// @Success 200 {object} DTO_Resp_Order "Обновлённая заявка"
// @Failure 400 {object} string "Invalid input or missing required fields"
// @Failure 401 {object} string "Unauthorized"
// @Failure 404 {object} string "Order not found"
// @Failure 500 {object} string "Internal server error"
// @Router /api/analysis_orders/{id}/resolve [put]
func (h *Handler) ApiResolveAnalysisOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid order ID: %v", err))
		return
	}

	var body struct {
		Action string `json:"action"` // "complete" | "reject"
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		if err.Error() == "EOF" {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("request body is empty, expected JSON with 'action' field"))
		} else {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err))
		}
		return
	}

	if body.Action != "complete" && body.Action != "reject" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("action must be 'complete' or 'reject'"))
		return
	}

	// Используем хардкод ID модератора, аналогично твоему QTask методу
	hardcodedModeratorID := uint(1)

	order, err := h.Repository.ResolveAnalysisOrder(uint(id), hardcodedModeratorID, body.Action)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Преобразуем экспертов в DTO
	var dtoExperts []DTO_Resp_OrderExpert
	for _, link := range order.ExpertsLinks {
		dtoExperts = append(dtoExperts, DTO_Resp_OrderExpert{
			ID_artcenter: link.ID_artcenter,
			ID_order:     link.ID_order,
			CenterX:      link.CenterX,
			CenterY:      link.CenterY,
			Title:        link.ArtExpert.Title,
			Name:         link.ArtExpert.Name,
			Algorithm:    link.ArtExpert.Algorithm,
		})
	}

	dtoOrder := DTO_Resp_Order{
		ID_order:      order.ID_order,
		OrderStatus:   order.OrderStatus,
		DateCreated:   order.DateCreated,
		DateFormed:    order.DateFormed,
		DateCompleted: order.DateCompleted,
		ResultX:       order.ResultX,
		ResultY:       order.ResultY,
		ID_creator:    order.ID_creator,
		ID_moderator:  order.ID_moderator,
		Experts:       dtoExperts,
	}

	ctx.JSON(http.StatusOK, dtoOrder)
}

// ApiDeleteAnalysisOrder удаляет заказ на анализ
// @Summary Удалить заказ на анализ
// @Description Полностью удаляет заказ анализа произведения искусства
// @Tags AnalysisOrders
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {object} DTO_Resp_SimpleID "ID удалённого заказа"
// @Failure 400 {object} string "Invalid order ID"
// @Failure 500 {object} string "Internal server error"
// @Router /api/analysis_orders/{id} [delete]
func (h *Handler) ApiDeleteAnalysisOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid order ID: %v", err))
		return
	}

	if err := h.Repository.DeleteAnalysisOrder(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_SimpleID{ID: id})
}

type updateExpertRequest struct {
	CenterX *float32 `json:"center_x"`
	CenterY *float32 `json:"center_y"`
}

// ApiRemoveExpertFromOrder удаляет эксперта из заказа
// @Summary Удалить эксперта из заказа
// @Description Удаляет связь между экспертом и заказом анализа
// @Tags M-M
// @Accept json
// @Produce json
// @Param order_id path int true "ID заказа"
// @Param expert_id path int true "ID эксперта"
// @Success 200 {object} DTO_Resp_OrderExpertLink "Информация об удалённой связи"
// @Failure 400 {object} string "Invalid IDs"
// @Failure 500 {object} string "Internal server error"
// @Router /api/orders/{order_id}/experts/{expert_id} [delete]
func (h *Handler) ApiRemoveExpertFromOrder(ctx *gin.Context) {
	orderID, err1 := strconv.Atoi(ctx.Param("order_id"))
	expertID, err2 := strconv.Atoi(ctx.Param("expert_id"))
	if err1 != nil || err2 != nil || orderID <= 0 || expertID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid ids"))
		return
	}

	if err := h.Repository.RemoveExpertFromOrder(uint(orderID), uint(expertID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_OrderExpertLink{OrderID: uint(orderID), ExpertID: int(expertID)})
}

// ApiUpdateExpertInOrder обновляет данные эксперта в заказе
// @Summary Обновить параметры эксперта
// @Description Обновляет координаты центра, переданные экспертом, в рамках заказа анализа
// @Tags M-M
// @Accept json
// @Produce json
// @Param order_id path int true "ID заказа"
// @Param expert_id path int true "ID эксперта"
// @Param request body updateExpertRequest true "Новые координаты эксперта"
// @Success 200 {object} DTO_Resp_UpdateExpert "Обновлённые данные эксперта"
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal server error"
// @Router /api/orders/{order_id}/experts/{expert_id} [put]
func (h *Handler) ApiUpdateExpertInOrder(ctx *gin.Context) {
	orderID, err1 := strconv.Atoi(ctx.Param("order_id"))
	expertID, err2 := strconv.Atoi(ctx.Param("expert_id"))
	if err1 != nil || err2 != nil || orderID <= 0 || expertID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid ids"))
		return
	}

	var req updateExpertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || (req.CenterX == nil && req.CenterY == nil) {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("center_x or center_y required"))
		return
	}

	if err := h.Repository.UpdateExpertInOrder(uint(orderID), uint(expertID), req.CenterX, req.CenterY); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_UpdateExpert{
		OrderID:  orderID,
		ExpertID: expertID,
		CenterX:  req.CenterX,
		CenterY:  req.CenterY,
	})
}
