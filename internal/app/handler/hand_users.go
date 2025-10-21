package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var creatorUserID uint = 1 // текущий пользователь (демо)

func (h *Handler) ApiRegisterUser(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RegisterUser(req.Login, req.Password); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	h.okJSON(ctx, http.StatusCreated, gin.H{"login": req.Login})
}

func (h *Handler) ApiGetMe(ctx *gin.Context) {
	user, err := h.Repository.GetUserByID(creatorUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, user)
}

func (h *Handler) ApiUpdateMe(ctx *gin.Context) {
	var req struct {
		Password *string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updated, err := h.Repository.UpdateUser(creatorUserID, req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	h.okJSON(ctx, http.StatusOK, updated)
}

func (h *Handler) ApiLogin(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.CheckUser(req.Login, req.Password); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	h.okJSON(ctx, http.StatusOK, gin.H{"login": req.Login})
}

func (h *Handler) ApiLogout(ctx *gin.Context) {
	h.okJSON(ctx, http.StatusOK, gin.H{"logout": true})
}
