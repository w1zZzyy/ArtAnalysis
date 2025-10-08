package handler

import (
	"errors"
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
		h.errorHandler(c, http.StatusNotFound, err)
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
