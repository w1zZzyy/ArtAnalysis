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
func (handler *Handler) RegisterHandler(r *gin.Engine) {
	r.GET("/experts", handler.GetArtExperts)
	r.GET("/expert/:id", handler.GetArtExpertByID)
	r.GET("/analysis_order/:id", handler.GetOrder)
	r.POST("/analysis_order/add/expert/:id_expert", handler.AddExpertToOrder)
	r.POST("/analysis_order/:order_id/delete", handler.DeleteOrder)

	// JSON API routes (exactly 21)

	// Experts (7)
	r.GET("/api/experts", handler.ApiExpertsList)                      // список экспертов с фильтрацией (по алгоритму)
	r.GET("/api/experts/:id", handler.ApiGetExpertByID)                // получить одного эксперта
	r.POST("/api/experts", handler.ApiAddExpert)                       // добавить нового эксперта
	r.PUT("/api/experts/:id", handler.ApiUpdateExpert)                 // изменить данные эксперта
	r.DELETE("/api/experts/:id", handler.ApiDeleteExpert)              // удалить эксперта (и изображение)
	r.POST("/api/draft/experts/:id", handler.ApiAddExpertToDraftOrder) // добавить эксперта в заявку-черновик
	r.POST("/api/experts/:id/image", handler.ApiUploadExpertImage)     // загрузить/заменить изображение эксперта (Minio)

	// Analysis orders (7)
	r.GET("/api/analysis_orders/current", handler.ApiGetCurrentDraftOrder)     // иконка корзины (черновик + кол-во экспертов)
	r.GET("/api/analysis_orders", handler.ApiListAnalysisOrders)               // список заявок с фильтрацией по дате и статусу
	r.GET("/api/analysis_orders/:id", handler.ApiGetAnalysisOrderByID)         // получить одну заявку (с экспертами)
	r.PUT("/api/analysis_orders/:id", handler.ApiUpdateAnalysisOrder)          // изменить поля заявки
	r.PUT("/api/analysis_orders/:id/form", handler.ApiFormAnalysisOrder)       // сформировать заявку (создателем)
	r.PUT("/api/analysis_orders/:id/resolve", handler.ApiResolveAnalysisOrder) // завершить/отклонить заявку (модератор)
	r.DELETE("/api/analysis_orders/:id", handler.ApiDeleteAnalysisOrder)       // логическое удаление заявки

	// m-m (2)
	r.DELETE("/api/orders/:order_id/experts/:expert_id", handler.ApiRemoveExpertFromOrder) // удалить эксперта из заявки
	r.PUT("/api/orders/:order_id/experts/:expert_id", handler.ApiUpdateExpertInOrder)      // изменить данные эксперта в заявке (например координаты)

	// Users (5)
	r.POST("/api/users/register", handler.ApiRegisterUser) // регистрация нового пользователя
	r.GET("/api/users/me", handler.ApiGetMe)               // данные текущего пользователя
	r.PUT("/api/users/me", handler.ApiUpdateMe)            // обновление данных пользователя
	r.POST("/api/auth/login", handler.ApiLogin)            // аутентификация
	r.POST("/api/auth/logout", handler.ApiLogout)          // деавторизация

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

func (h *Handler) okJSON(ctx *gin.Context, statusCode int, payload interface{}) {
	ctx.JSON(statusCode, gin.H{
		"status": "ok",
		"data":   payload,
	})
}
