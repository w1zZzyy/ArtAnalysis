package handler

import (
	"github.com/w1zZzyy22/art-analysis/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler регистрирует маршруты для работы с экспертами и заявками
func (h *Handler) RegisterHandler(r *gin.Engine) {
	r.GET("/experts", h.GetArtExperts)
	r.GET("/expert/:id", h.GetArtExpertByID)

	r.GET("/analysis_order/:id", h.GetOrder)
	r.POST("/analysis_order/add/expert/:id_expert", h.AddExpertToOrder) // ORM
	r.POST("/analysis_order/:order_id/delete", h.DeleteOrder)           // логическое удаление заявки
}

// RegisterStatic регистрирует статику и шаблоны
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

// errorHandler удобный вывод ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
