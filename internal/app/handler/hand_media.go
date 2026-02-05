package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ApiGetExpertMedia godoc
// @Summary      Получить медиафайлы эксперта
// @Description  Возвращает список всех медиафайлов (фото/видео) эксперта, отсортированных по ID.
// @Tags         ExpertMedia
// @Produce      json
// @Param        id path int true "ID эксперта"
// @Success      200 {array} DTO_Resp_ExpertMedia
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/experts/{id}/media [get]
func (h *Handler) ApiGetExpertMedia(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	media, err := h.Repository.GetMediaByExpertID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var response []DTO_Resp_ExpertMedia
	for _, m := range media {
		response = append(response, DTO_Resp_ExpertMedia{
			ID_media:     m.ID_media,
			ID_artcenter: m.ID_artcenter,
			MediaURL:     m.MediaURL,
			MediaType:    m.MediaType,
			CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	if response == nil {
		response = []DTO_Resp_ExpertMedia{}
	}

	ctx.JSON(http.StatusOK, response)
}

// ApiUploadExpertMedia godoc
// @Summary      Загрузить медиафайл эксперта
// @Description  Загружает фото или видео для эксперта в MinIO и сохраняет запись в БД.
// @Tags         ExpertMedia
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path int true "ID эксперта"
// @Param        file formData file true "Медиафайл (фото/видео)"
// @Success      201 {object} DTO_Resp_ExpertMedia
// @Failure      400 {object} map[string]string "Некорректный ID или отсутствует файл"
// @Failure      500 {object} map[string]string "Ошибка при загрузке"
// @Router       /api/experts/{id}/media [post]
func (h *Handler) ApiUploadExpertMedia(ctx *gin.Context) {
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

	media, err := h.Repository.AddExpertMedia(ctx, uint(id), fileHeader)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, DTO_Resp_ExpertMedia{
		ID_media:     media.ID_media,
		ID_artcenter: media.ID_artcenter,
		MediaURL:     media.MediaURL,
		MediaType:    media.MediaType,
		CreatedAt:    media.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ApiDeleteExpertMedia godoc
// @Summary      Удалить медиафайл
// @Description  Удаляет конкретный медиафайл эксперта по ID.
// @Tags         ExpertMedia
// @Produce      json
// @Param        id       path int true "ID эксперта"
// @Param        media_id path int true "ID медиафайла"
// @Success      200 {object} DTO_Resp_SimpleID
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      404 {object} map[string]string "Медиафайл не найден"
// @Failure      500 {object} map[string]string "Ошибка при удалении"
// @Router       /api/experts/{id}/media/{media_id} [delete]
func (h *Handler) ApiDeleteExpertMedia(ctx *gin.Context) {
	expertID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || expertID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	mediaID, err := strconv.Atoi(ctx.Param("media_id"))
	if err != nil || mediaID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем, что медиафайл принадлежит данному эксперту
	media, err := h.Repository.GetMediaByID(uint(mediaID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if media.ID_artcenter != uint(expertID) {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteExpertMedia(ctx, uint(mediaID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, DTO_Resp_SimpleID{ID: mediaID})
}

// ApiGetExpertWithMedia godoc
// @Summary      Получить эксперта с медиафайлами
// @Description  Возвращает полную информацию об эксперте включая все медиафайлы.
// @Tags         Experts
// @Produce      json
// @Param        id path int true "ID эксперта"
// @Success      200 {object} DTO_Resp_ExpertWithMedia
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      404 {object} map[string]string "Эксперт не найден"
// @Router       /api/experts/{id}/full [get]
func (h *Handler) ApiGetExpertWithMedia(ctx *gin.Context) {
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

	media, err := h.Repository.GetMediaByExpertID(uint(id))
	if err != nil {
		media = nil
	}

	var mediaResp []DTO_Resp_ExpertMedia
	for _, m := range media {
		mediaResp = append(mediaResp, DTO_Resp_ExpertMedia{
			ID_media:     m.ID_media,
			ID_artcenter: m.ID_artcenter,
			MediaURL:     m.MediaURL,
			MediaType:    m.MediaType,
			CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	if mediaResp == nil {
		mediaResp = []DTO_Resp_ExpertMedia{}
	}

	ctx.JSON(http.StatusOK, DTO_Resp_ExpertWithMedia{
		ID_artcenter: expert.ID_artcenter,
		Title:        expert.Title,
		Description:  expert.Description,
		Status:       expert.Status,
		Name:         expert.Name,
		Algorithm:    expert.Algorithm,
		ImgURL:       expert.ImgURL,
		Media:        mediaResp,
	})
}
