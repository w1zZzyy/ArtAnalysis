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
	r.POST("/login", handler.Login)
	r.POST("/users", handler.Register)

	// experts
	r.GET("/api/experts", handler.ApiExpertsList)
	r.GET("/api/experts/:id", handler.ApiGetExpertByID)
	r.GET("/api/experts/:id/full", handler.ApiGetExpertWithMedia)
	r.GET("/api/experts/:id/media", handler.ApiGetExpertMedia)

	// HTML
	r.GET("/ArtAnalysis", handler.GetArtExperts)
	r.GET("/expert_properties/:id", handler.GetArtExpertByID)

	// Internal эндпоинты для межсервисного взаимодействия (с псевдо-авторизацией по ключу)
	r.POST("/api/internal/analysis-result", handler.ApiReceiveAnalysisResult)
	r.PUT("/api/internal/analysis-result/:id", handler.ApiUpdateAnalysisResultManual)

	// Эндпоинты, доступные только модераторам
	moderator := r.Group("/")
	moderator.Use(handler.AuthMiddleware, handler.ModeratorMiddleware)
	{
		// Experts
		moderator.POST("/api/experts", handler.ApiAddExpert)
		moderator.PUT("/api/experts/:id", handler.ApiUpdateExpert)
		moderator.DELETE("/api/experts/:id", handler.ApiDeleteExpert)
		moderator.POST("/api/experts/:id/image", handler.ApiUploadExpertImage)

		// Expert Media
		moderator.POST("/api/experts/:id/media", handler.ApiUploadExpertMedia)
		moderator.DELETE("/api/experts/:id/media/:media_id", handler.ApiDeleteExpertMedia)

		// Center Request
		moderator.PUT("/api/center_request/:id/resolve", handler.ApiResolveCenterRequest)

		// Async analysis (только модератор может запускать анализ)
		moderator.POST("/api/center_request/:id/analyze", handler.ApiStartAsyncAnalysis)
	}

	// Эндпоинты, доступные всем авторизованным пользователям
	auth := r.Group("/")
	auth.Use(handler.AuthMiddleware)
	{
		// Users
		auth.POST("/api/auth/logout", handler.Logout)
		auth.GET("/api/users/me", handler.ApiMe)
		auth.PUT("/api/users/me", handler.ApiUpdateMe)

		// Связи
		auth.DELETE("/api/center_request/:id/experts/:expert_id", handler.ApiRemoveExpertFromCenterRequest)
		auth.PUT("/api/center_request/:id/experts/:expert_id", handler.ApiUpdateExpertInCenterRequest)

		// HTML
		auth.GET("/center_request/:id", handler.GetCenterRequest)
		auth.POST("/center_request/add/expert/:id_expert", handler.AddExpertToRequest)
		auth.POST("/center_request/:id/delete", handler.DeleteCenterRequest)

		// Experts
		auth.POST("/api/draft/experts/:id", handler.ApiAddExpertToDraftCenterRequest)

		// Center Request
		auth.GET("/api/center_request/current", handler.ApiGetCurrCenterRequest)
		auth.GET("/api/center_request", handler.ApiListCenterRequest)
		auth.GET("/api/center_request/:id", handler.ApiGetCenterRequestByID)
		auth.PUT("/api/center_request/:id", handler.ApiUpdateCenterRequest)
		auth.PUT("/api/center_request/:id/form", handler.ApiFormCenterRequest)
		auth.DELETE("/api/center_request/:id", handler.ApiDeleteCenterRequest)
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
