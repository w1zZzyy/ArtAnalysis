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

func (h *Handler) ApiListAnalysisOrders(ctx *gin.Context) {
	status := ctx.Query("status")
	from := ctx.Query("from")
	to := ctx.Query("to")
	orders, err := h.Repository.ListAnalysisOrders(status, from, to)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, gin.H{"items": orders})
}

func (h *Handler) ApiGetAnalysisOrderByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	order, err := h.Repository.GetAnalysisOrderByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, order)
}

func (h *Handler) ApiUpdateAnalysisOrder(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var body struct {
		Status      *string `json:"status"`
		ModeratorID *uint   `json:"moderator_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	updated, err := h.Repository.UpdateAnalysisOrder(uint(id), body.Status, body.ModeratorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, updated)
}

func (h *Handler) ApiFormAnalysisOrder(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	formed, err := h.Repository.FormAnalysisOrder(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, formed)
}

func (h *Handler) ApiResolveAnalysisOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid order id: %v", err))
		return
	}

	var body struct {
		Action string `json:"action"` // "complete" | "reject"
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if body.Action != "complete" && body.Action != "reject" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("action must be 'complete' or 'reject'"))
		return
	}

	// используем хардкод для модератора, как в оригинале
	hardcodedModeratorID := uint(1)

	resolved, err := h.Repository.ResolveAnalysisOrder(uint(id), hardcodedModeratorID, body.Action)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	h.okJSON(ctx, http.StatusOK, resolved)
}

func (h *Handler) ApiDeleteAnalysisOrder(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteAnalysisOrder(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, gin.H{"id": id})
}

type updateExpertRequest struct {
	CenterX *float32 `json:"center_x"`
	CenterY *float32 `json:"center_y"`
}

// Удаление эксперта из заявки
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

	h.okJSON(ctx, http.StatusOK, gin.H{"order_id": orderID, "expert_id": expertID})
}

// Обновление координат эксперта в заявке
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

	h.okJSON(ctx, http.StatusOK, gin.H{
		"order_id":  orderID,
		"expert_id": expertID,
		"center_x":  req.CenterX,
		"center_y":  req.CenterY,
	})
}
