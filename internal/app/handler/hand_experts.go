package handler

import (
	"context"

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

type expertCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Algorithm   string `json:"algorithm"`
	Status      *bool  `json:"status"`
}

func (h *Handler) ApiExpertsList(ctx *gin.Context) {
	title := ctx.Query("title")
	experts, err := h.Repository.ListExperts(title)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, gin.H{"items": experts})
}

func (h *Handler) ApiGetExpertByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	expert, err := h.Repository.GetExpertByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, expert)
}

func (h *Handler) ApiAddExpert(ctx *gin.Context) {
	var req expertCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	expert := model.ArtExpert{
		Title:       req.Title,
		Description: req.Description,
		Name:        req.Name,
		Algorithm:   req.Algorithm,
	}
	if req.Status != nil {
		expert.Status = *req.Status
	}

	if err := h.Repository.AddExpert(&expert); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusCreated, expert)
}

func (h *Handler) ApiUpdateExpert(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req expertCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	updated, err := h.Repository.UpdateExpert(uint(id), req.Title, req.Description, req.Name, req.Algorithm, req.Status)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, updated)
}

func (h *Handler) ApiDeleteExpert(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteExpert(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, gin.H{"id": id})
}

func (h *Handler) ApiAddExpertToDraftOrder(ctx *gin.Context) {
	expertID, _ := strconv.Atoi(ctx.Param("id"))
	var body struct {
		OrderID uint `json:"order_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.AddExpertToDraftOrder(uint(expertID), body.OrderID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusCreated, gin.H{"expert_id": expertID, "order_id": body.OrderID})
}

func (h *Handler) ApiUploadExpertImage(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	file, err := ctx.FormFile("file")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	url, err := h.Repository.SaveExpertImage(context.Background(), uint(id), file)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusCreated, gin.H{"id": id, "image": url})
}
