package handler

import (
	"github.com/w1zZzyy22/art-analysis/internal/app/config"
	appredis "github.com/w1zZzyy22/art-analysis/internal/app/redis"
	"github.com/w1zZzyy22/art-analysis/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler структура обработчиков API
// @Description Основная структура содержащая зависимости обработчиков
type Handler struct {
	Repository *repository.Repository
	Redis      *appredis.Client
	JWTConfig  *config.JWTConfig
}

// NewHandler создает новый экземпляр Handler
// @Description Конструктор для создания экземпляра Handler с зависимостями
func NewHandler(r *repository.Repository, redis *appredis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redis,
		JWTConfig:  jwtConfig,
	}
}

// RegisterHandler регистрирует все маршруты API
// @Description Функция для регистрации всех маршрутов приложения с группировкой по правам доступа
func (handler *Handler) RegisterHandler(r *gin.Engine) {
	// Публичные эндпоинты (без авторизации)
	r.POST("/api/auth/login", handler.Login)
	r.POST("/api/users/register", handler.Register)

	// experts
	r.GET("/expert/:id", handler.GetArtExpertByID)
	r.GET("/analysis_order/:id", handler.GetOrder)

	// HTML
	r.GET("/experts", handler.GetArtExperts)

	// Эндпоинты, доступные только модераторам
	moderator := r.Group("/")
	moderator.Use(handler.AuthMiddleware, handler.ModeratorMiddleware)
	{
		// Experts
		moderator.POST("/api/experts", handler.ApiAddExpert)
		moderator.PUT("/api/experts/:id", handler.ApiUpdateExpert)
		moderator.DELETE("/api/experts/:id", handler.ApiDeleteExpert)
		moderator.POST("/api/experts/:id/image", handler.ApiUploadExpertImage)

		// Analysis orders
		moderator.PUT("/api/analysis_orders/:id/resolve", handler.ApiResolveAnalysisOrder)
		moderator.DELETE("/api/analysis_orders/:id", handler.ApiDeleteAnalysisOrder)
	}

	// Эндпоинты, доступные всем авторизованным пользователям
	auth := r.Group("/")
	auth.Use(handler.AuthMiddleware)
	{
		// Users
		auth.POST("/api/auth/logout", handler.Logout)
		auth.GET("/api/users/me", handler.ApiMe)
		auth.PUT("/api/users/me", handler.ApiUpdateMe)

		// Experts (только просмотр и черновики)
		auth.GET("/api/experts", handler.ApiExpertsList)
		auth.GET("/api/experts/:id", handler.ApiGetExpertByID)
		auth.POST("/api/draft/experts/:id", handler.ApiAddExpertToDraftOrder)

		// Analysis orders (создание, редактирование, просмотр)
		auth.GET("/api/analysis_orders/current", handler.ApiGetCurrentDraftOrder)
		auth.GET("/api/analysis_orders", handler.ApiListAnalysisOrders)
		auth.GET("/api/analysis_orders/:id", handler.ApiGetAnalysisOrderByID)
		auth.PUT("/api/analysis_orders/:id", handler.ApiUpdateAnalysisOrder)
		auth.PUT("/api/analysis_orders/:id/form", handler.ApiFormAnalysisOrder)

		// Many-to-Many: заявки ↔ эксперты
		auth.DELETE("/api/orders/:order_id/experts/:expert_id", handler.ApiRemoveExpertFromOrder)
		auth.PUT("/api/orders/:order_id/experts/:expert_id", handler.ApiUpdateExpertInOrder)

		// HTML
		auth.POST("/analysis_order/add/expert/:id_expert", handler.AddExpertToOrder)
		auth.POST("/analysis_order/:order_id/delete", handler.DeleteOrder)
	}

}

// RegisterStatic регистрирует статические файлы и шаблоны
// @Description Настраивает обслуживание статических файлов и HTML шаблонов
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

// errorHandler - внутренний вспомогательный метод для обработки ошибок
// Не экспортируется в Swagger документацию
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

// okJSON - внутренний вспомогательный метод для успешных ответов
// Не экспортируется в Swagger документацию
func (h *Handler) okJSON(ctx *gin.Context, statusCode int, payload interface{}) {
	ctx.JSON(statusCode, gin.H{
		"status": "ok",
		"data":   payload,
	})
}
