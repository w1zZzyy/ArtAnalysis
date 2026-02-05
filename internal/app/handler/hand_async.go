package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Секретный ключ для псевдо-авторизации от async сервиса (8 байт = 16 hex символов)
const INTERNAL_SERVICE_KEY = "a1b2c3d4e5f67890"

// URL асинхронного сервиса
const ASYNC_SERVICE_URL = "http://localhost:8090"

// ApiReceiveAnalysisResult получает результат от асинхронного сервиса
// @Summary Получить результат асинхронного анализа
// @Description Эндпоинт для получения результатов от асинхронного Python сервиса. Требует псевдо-авторизацию через service_key.
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body DTO_Req_AnalysisResult true "Результат анализа"
// @Success 200 {object} map[string]interface{} "Результат успешно принят"
// @Failure 400 {object} string "Invalid input"
// @Failure 401 {object} string "Unauthorized - неверный service_key"
// @Failure 404 {object} string "Request not found"
// @Failure 500 {object} string "Internal server error"
// @Router /api/internal/analysis-result [post]
func (h *Handler) ApiReceiveAnalysisResult(ctx *gin.Context) {
	var req DTO_Req_AnalysisResult

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Проверка псевдо-авторизации
	if req.ServiceKey != INTERNAL_SERVICE_KEY {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("invalid service_key"))
		return
	}

	if req.RequestID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid request_id"))
		return
	}

	// Сохраняем результат в БД
	err := h.Repository.UpdateAnalysisResult(
		uint(req.RequestID),
		req.Success,
		req.AnalysisResult,
		req.ConfidenceScore,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to save result: %v", err))
		return
	}

	logrus.Infof("[ASYNC RESULT] Получен результат для заявки %d: success=%v, result=%v",
		req.RequestID, req.Success, req.AnalysisResult)

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "accepted",
		"request_id": req.RequestID,
		"message":    "Результат анализа успешно сохранён",
	})
}

// ApiStartAsyncAnalysis запускает асинхронный анализ
// @Summary Запустить асинхронный анализ
// @Description Отправляет заявку в асинхронный Python сервис для анализа. Результат будет получен через 5-10 секунд.
// @Tags CenterRequest
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]interface{} "Анализ запущен"
// @Failure 400 {object} string "Invalid request ID"
// @Failure 401 {object} string "Unauthorized"
// @Failure 404 {object} string "Request not found"
// @Failure 500 {object} string "Internal server error"
// @Security BearerAuth
// @Router /api/center_request/{id}/analyze [post]
func (h *Handler) ApiStartAsyncAnalysis(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid request ID"))
		return
	}

	// Получаем заявку
	request, err := h.Repository.GetRequestWithExperts(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Формируем запрос к async сервису
	asyncPayload := map[string]interface{}{
		"request_id":  request.ID_request,
		"description": request.RequestDescription,
	}
	if request.FactorX != nil {
		asyncPayload["factor_x"] = *request.FactorX
	}
	if request.FactorY != nil {
		asyncPayload["factor_y"] = *request.FactorY
	}

	jsonData, err := json.Marshal(asyncPayload)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to marshal request: %v", err))
		return
	}

	// Отправляем запрос в async сервис
	resp, err := http.Post(
		ASYNC_SERVICE_URL+"/api/analyze",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		logrus.Errorf("Failed to call async service: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("async service unavailable: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("async service returned status: %d", resp.StatusCode))
		return
	}

	// Сбрасываем предыдущий результат анализа
	h.Repository.ClearAnalysisResult(uint(id))

	logrus.Infof("[ASYNC] Запущен анализ для заявки %d", id)

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "accepted",
		"request_id": id,
		"message":    "Асинхронный анализ запущен. Результат будет получен через 5-10 секунд.",
	})
}

// ApiUpdateAnalysisResult позволяет вручную обновить результат анализа (с ключом)
// @Summary Обновить результат анализа вручную
// @Description Позволяет обновить результат анализа заявки вручную. Требует service_key для авторизации.
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body object true "Новый результат анализа"
// @Success 200 {object} map[string]interface{} "Результат обновлён"
// @Failure 400 {object} string "Invalid input"
// @Failure 401 {object} string "Unauthorized"
// @Failure 500 {object} string "Internal server error"
// @Router /api/internal/analysis-result/{id} [put]
func (h *Handler) ApiUpdateAnalysisResultManual(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid request ID"))
		return
	}

	var req struct {
		AnalysisResult  *string  `json:"analysis_result"`
		ConfidenceScore *float32 `json:"confidence_score"`
		Success         *bool    `json:"success"`
		ServiceKey      string   `json:"service_key" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Проверка псевдо-авторизации
	if req.ServiceKey != INTERNAL_SERVICE_KEY {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("invalid service_key"))
		return
	}

	success := true
	if req.Success != nil {
		success = *req.Success
	}

	err = h.Repository.UpdateAnalysisResult(
		uint(id),
		success,
		req.AnalysisResult,
		req.ConfidenceScore,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to update result: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "updated",
		"request_id": id,
		"message":    "Результат анализа обновлён",
	})
}
