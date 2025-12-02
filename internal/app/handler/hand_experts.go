package handler

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"time"

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

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("требуется авторизация"))
		return
	}

	// Получаем черновик заявки для текущего пользователя
	draftRequest, _ := h.Repository.GetDraftRequest(userID)
	var requestID uint = 0
	var expertsCount int = 0

	if draftRequest != nil {
		fullrequest, err := h.Repository.GetRequestWithExperts(draftRequest.ID_request)
		if err == nil {
			requestID = fullrequest.ID_request
			expertsCount = len(fullrequest.ExpertsLinks)
		}
	}

	ctx.HTML(http.StatusOK, "experts_list.html", gin.H{
		"experts":         experts,
		"expertSearching": search,
		"requestID":       requestID,
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

// ---- JSON API (services/gates) ----

// ApiExpertsList godoc
// @Summary      Получить список экспертов
// @Description  Возвращает список всех экспертов. Поддерживает фильтрацию по названию.
// @Tags         Experts
// @Produce      json
// @Param        title query string false "Фильтр по названию эксперта (подстрока)"
// @Success      200 {array} DTO_Resp_Expert
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/experts [get]
func (h *Handler) ApiExpertsList(ctx *gin.Context) {
	title := ctx.Query("title")

	experts, err := h.Repository.ListExperts(title)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var representExperts []DTO_Resp_Expert
	for _, expert := range experts {
		representExperts = append(representExperts, DTO_Resp_Expert{
			ID_artcenter: expert.ID_artcenter,
			Title:        expert.Title,
			Description:  expert.Description,
			Status:       expert.Status,
			Name:         expert.Name,
			Algorithm:    expert.Algorithm,
			ImgURL:       expert.ImgURL,
		})
	}

	ctx.JSON(http.StatusOK, representExperts)
}

// ApiGetExpertByID godoc
// @Summary      Получить эксперта по ID
// @Description  Возвращает полную информацию об эксперте по его идентификатору.
// @Tags         Experts
// @Produce      json
// @Param        id path int true "ID эксперта"
// @Success      200 {object} model.ArtExpert
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      404 {object} map[string]string "Эксперт не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/experts/{id} [get]
func (h *Handler) ApiGetExpertByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	expert, err := h.Repository.GetExpertByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, expert)
}

// ApiAddExpert godoc
// @Summary      Создать нового эксперта
// @Description  Создаёт нового эксперта с указанными атрибутами.
// @Tags         Experts
// @Accept       json
// @Produce      json
// @Param        expert body DTO_Req_ExpertCreate true "Данные нового эксперта"
// @Success      201 {object} model.ArtExpert
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      500 {object} map[string]string "Ошибка при создании эксперта"
// @Router       /api/experts [post]
func (h *Handler) ApiAddExpert(ctx *gin.Context) {
	var req DTO_Req_ExpertCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if req.Title == "" || req.Description == "" || req.Name == "" || req.Algorithm == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("missing required fields"))
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

	ctx.JSON(http.StatusCreated, expert)
}

// ApiUpdateExpert godoc
// @Summary      Обновить эксперта
// @Description  Обновляет параметры существующего эксперта.
// @Tags         Experts
// @Accept       json
// @Produce      json
// @Param        id     path int true "ID эксперта"
// @Param        expert body DTO_Req_ExpertCreate true "Обновлённые данные эксперта"
// @Success      200 {object} model.ArtExpert
// @Failure      400 {object} map[string]string "Некорректные данные или ID"
// @Failure      500 {object} map[string]string "Ошибка при обновлении эксперта"
// @Router       /api/experts/{id} [put]
func (h *Handler) ApiUpdateExpert(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req DTO_Req_ExpertCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updated, err := h.Repository.UpdateExpert(
		uint(id),
		req.Title,
		req.Description,
		req.Name,
		req.Algorithm,
		req.Status,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

// ApiDeleteExpert godoc
// @Summary      Удалить эксперта
// @Description  Полностью удаляет эксперта по ID.
// @Tags         Experts
// @Produce      json
// @Param        id path int true "ID эксперта"
// @Success      200 {object} DTO_Resp_SimpleID
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      500 {object} map[string]string "Ошибка при удалении эксперта"
// @Router       /api/experts/{id} [delete]
func (h *Handler) ApiDeleteExpert(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteExpert(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_SimpleID{ID: id})
}

// ApiAddExpertToDraftCenterRequest godoc
// @Summary      Добавить эксперта в черновик заявки
// @Description  Добавляет указанного эксперта в текущий черновой заказ пользователя.
// @Tags         Experts
// @Produce      json
// @Param        id path int true "ID эксперта"
// @Success      201 {object} DTO_Resp_CenterRequestExpertLink
// @Failure      400 {object} map[string]string "Некорректный ID эксперта"
// @Failure      401 {object} map[string]string "Требуется авторизация"
// @Failure      500 {object} map[string]string "Ошибка при добавлении эксперта"
// @Router       /api/draft/experts/{id} [post]
func (h *Handler) ApiAddExpertToDraftCenterRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	request, err := h.Repository.GetDraftRequest(userID)
	if err != nil {
		newReq := model.CenterRequest{
			ID_creator:    userID,
			RequestStatus: model.StatusDraft,
			DateCreated:   time.Now(),
		}
		if createErr := h.Repository.CreateRequest(&newReq); createErr != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, createErr)
			return
		}
		request = &newReq
	}

	if err := h.Repository.AddExpertToRequest(request.ID_request, uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusCreated, DTO_Resp_CenterRequestExpertLink{
		RequestID: request.ID_request,
		ExpertID:  id,
	})
}

// ApiGetCurrCenterRequest godoc
// @Summary      Получить информацию о текущем черновом заказе
// @Description  Возвращает ID текущего чернового заказа и количество добавленных экспертов.
// @Tags         CenterRequest
// @Produce      json
// @Success      200 {object} DTO_Resp_CurrCenterRequestInfo
// @Failure      401 {object} map[string]string "Требуется авторизация"
// @Failure      500 {object} map[string]string "Ошибка при получении данных"
// @Router       /api/center_request/current [get]
func (h *Handler) ApiGetCurrCenterRequest(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("требуется авторизация"))
		return
	}
	draftTask, _ := h.Repository.GetDraftRequest(userID)
	var requestID uint = 0
	var expertsCount int = 0
	if draftTask != nil {
		fullTask, err := h.Repository.GetRequestWithExperts(draftTask.ID_request)
		if err == nil {
			requestID = fullTask.ID_request
			expertsCount = len(fullTask.ExpertsLinks)
		}
	}
	ctx.JSON(http.StatusOK, DTO_Resp_CurrCenterRequestInfo{RequestID: requestID, ExpertsCount: expertsCount})
}

// ApiUploadExpertImage godoc
// @Summary      Загрузить изображение эксперта
// @Description  Загружает изображение и сохраняет URL в базе данных.
// @Tags         Experts
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path int true "ID эксперта"
// @Param        file formData file true "Изображение эксперта"
// @Success      201 {object} DTO_Resp_UploadImg
// @Failure      400 {object} map[string]string "Некорректный ID или отсутствует файл"
// @Failure      500 {object} map[string]string "Ошибка при загрузке изображения"
// @Router       /api/experts/{id}/image [post]
func (h *Handler) ApiUploadExpertImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Загрузка изображения и обновление записи в базе
	imageURL, err := h.Repository.SaveExpertImage(ctx, uint(id), fileHeader)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, DTO_Resp_UploadImg{
		ID:    id,
		Image: imageURL,
	})
}

func generateSafeImageName(fh *multipart.FileHeader) string {
	base := fh.Filename
	ext := filepath.Ext(base)
	name := base
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	slug := re.ReplaceAllString(name, "-")
	if slug == "" {
		slug = "image"
	}
	if ext == "" {
		ext = ".png"
	}
	return slug + ext
}
