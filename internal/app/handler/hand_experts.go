package handler

import (
	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetArtExperts(ctx *gin.Context) {
	var experts []model.ArtExpert
	var err error

	search := ctx.Query("expertSearching")
	if search == "" {
		experts, err = h.Repository.GetExperts()
	} else {
		experts, err = h.Repository.GetExpertsByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Получаем черновик заявки для текущего пользователя
	draftOrder, _ := h.Repository.GetDraftOrder(hardcodedUserID)
	var orderID uint = 0
	var expertsCount int = 0

	if draftOrder != nil {
		fullOrder, err := h.Repository.GetOrderWithExperts(draftOrder.ID_order)
		if err == nil {
			orderID = fullOrder.ID_order
			expertsCount = len(fullOrder.ExpertsLinks)
		}
	}

	ctx.HTML(http.StatusOK, "experts_list.html", gin.H{
		"experts":         experts,
		"expertSearching": search,
		"orderID":         orderID,
		"expertsCount":    expertsCount,
	})
}

func (h *Handler) GetArtExpertByID(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	expert, err := h.Repository.GetExpertByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "expert_properties.html", expert)
}
