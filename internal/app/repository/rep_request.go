package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"gorm.io/gorm"
)

func (r *Repository) GetDraftRequest(userID uint) (*model.CenterRequest, error) {
	var request model.CenterRequest

	err := r.db.Where("id_creator = ? AND request_status = ?", userID, model.StatusDraft).
		First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not an error, just no draft exists
		}
		return nil, err
	}
	return &request, nil
}

func (r *Repository) CreateRequest(request *model.CenterRequest) error {
	if request.DateCreated.IsZero() {
		request.DateCreated = time.Now()
	}
	return r.db.Create(request).Error
}

func (r *Repository) AddExpertToRequest(requestID, expertID uint) error {
	var count int64
	r.db.Model(&model.ExpertsToRequest{}).
		Where("id_request = ? AND id_artcenter = ?", requestID, expertID).
		Count(&count)
	if count > 0 {
		return errors.New("expert already linked to this request")
	}

	link := model.ExpertsToRequest{
		ID_request:   requestID,
		ID_artcenter: expertID,
	}
	return r.db.Create(&link).Error
}

func (r *Repository) GetRequestWithExperts(requestID uint) (*model.CenterRequest, error) {
	var request model.CenterRequest

	err := r.db.
		Preload("ExpertsLinks.ArtExpert"). // Загружаем через связующую таблицу
		First(&request, requestID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("request not found")
		}
		return nil, fmt.Errorf("error fetching request: %v", err)
	}

	// Проверяем, что заявка не удалена
	if request.RequestStatus == model.StatusDeleted {
		return nil, errors.New("request not found or has been deleted")
	}

	return &request, nil
}

// LogicallyDeleteRequest выполняет логическое удаление заявки через SQL UPDATE.
func (r *Repository) LogicallyDeleteRequest(requestID uint) error {
	result := r.db.Exec("UPDATE center_requests SET request_status = ? WHERE id_request = ?", model.StatusDeleted, requestID)
	return result.Error
}

func (r *Repository) GetCurrentDraftRequest(userID uint) (*model.CenterRequest, int, error) {
	var request model.CenterRequest
	err := r.db.Preload("ExpertsLinks").Where("id_creator = ? AND request_status = ?", userID, "черновик").First(&request).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}
	count := len(request.ExpertsLinks)
	return &request, count, nil
}

func (r *Repository) ListCenterRequests(status, from, to string) ([]*model.CenterRequest, error) {
	var requests []*model.CenterRequest
	q := r.db.Preload("ExpertsLinks.ArtExpert").Model(&model.CenterRequest{})
	if status != "" {
		q = q.Where("request_status = ?", status)
	} else {
		q = q.Where("request_status NOT IN ?", []string{model.StatusDeleted, model.StatusDraft})
	}
	if from != "" {
		q = q.Where("date_created >= ?", from)
	}
	if to != "" {
		q = q.Where("date_created <= ?", to)
	}
	if err := q.Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *Repository) ListRequestsByUser(userID uint, status, from, to string) ([]*model.CenterRequest, error) {
	var requests []*model.CenterRequest

	query := r.db.
		Preload("ExpertsLinks.ArtExpert"). // Исправлено: было ExpertsToRequest
		Preload("User").                   // Загрузка данных пользователя-создателя
		Preload("Moderator").              // Загрузка данных модератора
		Where("id_creator = ?", userID)    // Фильтруем по создателю заявки

	if status != "" {
		query = query.Where("request_status = ?", status)
	} else {
		query = query.Where("request_status NOT IN ?", []string{model.StatusDeleted, model.StatusDraft})
	}
	if from != "" {
		query = query.Where("date_created >= ?", from)
	}
	if to != "" {
		query = query.Where("date_created <= ?", to)
	}

	err := query.Order("date_created DESC").Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *Repository) GetCenterRequestByID(id uint) (*model.CenterRequest, error) {
	var request model.CenterRequest
	if err := r.db.Preload("ExpertsLinks.ArtExpert").Preload("User").First(&request, id).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *Repository) UpdateCenterRequest(id uint, description string) (*model.CenterRequest, error) {
	var request model.CenterRequest
	if err := r.db.First(&request, id).Error; err != nil {
		return nil, err
	}
	request.RequestDescription = description
	if err := r.db.Save(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *Repository) FormCenterRequest(id uint) (*model.CenterRequest, error) {
	var request model.CenterRequest
	if err := r.db.First(&request, id).Error; err != nil {
		return nil, err
	}
	if request.RequestStatus != model.StatusDraft {
		return nil, errors.New("only draft requests can be formed")
	}
	request.RequestStatus = model.StatusFormed
	now := time.Now()
	request.DateFormed = &now
	if err := r.db.Save(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *Repository) ResolveCenterRequest(id uint, moderatorID uint, action string, resolvedAt time.Time) (*model.CenterRequest, error) {
	var request model.CenterRequest
	if err := r.db.First(&request, id).Error; err != nil {
		return nil, err
	}
	if request.RequestStatus != model.StatusFormed {
		return nil, errors.New("only formed requests can be resolved")
	}

	var newStatus string
	if action == "complete" {
		newStatus = model.StatusCompleted
	} else if action == "reject" {
		newStatus = model.StatusRejected
	} else {
		return nil, errors.New("invalid action")
	}

	// Обновляем только нужные поля, чтобы не перезаписать FactorX/FactorY
	if err := r.db.Model(&request).Updates(map[string]interface{}{
		"request_status":  newStatus,
		"id_moderator":    moderatorID,
		"date_conclusion": resolvedAt,
	}).Error; err != nil {
		return nil, err
	}

	// Перезагружаем заявку с актуальными данными
	if err := r.db.Preload("ExpertsLinks").First(&request, id).Error; err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *Repository) DeleteCenterRequest(id uint) error {
	return r.db.Model(&model.CenterRequest{}).Where("id_request = ?", id).Update("request_status", model.StatusDeleted).Error
}

func (r *Repository) RemoveExpertFromRequest(requestID, expertID uint) error {
	return r.db.Where("id_request = ? AND id_artcenter = ?", requestID, expertID).Delete(&model.ExpertsToRequest{}).Error
}

func (r *Repository) UpdateExpertInRequest(requestID, expertID uint, centerX, centerY *float32) error {
	var link model.ExpertsToRequest
	if err := r.db.First(&link, "id_request = ? AND id_artcenter = ?", requestID, expertID).Error; err != nil {
		return err
	}
	if centerX != nil {
		link.CenterX = centerX
	}
	if centerY != nil {
		link.CenterY = centerY
	}
	return r.db.Save(&link).Error
}

// UpdateAnalysisResult обновляет результат асинхронного анализа
func (r *Repository) UpdateAnalysisResult(requestID uint, success bool, result *string, confidence *float32) error {
	return r.db.Model(&model.CenterRequest{}).
		Where("id_request = ?", requestID).
		Updates(map[string]interface{}{
			"analysis_success": success,
			"analysis_result":  result,
			"confidence_score": confidence,
		}).Error
}

// ClearAnalysisResult сбрасывает результат анализа
func (r *Repository) ClearAnalysisResult(requestID uint) error {
	return r.db.Model(&model.CenterRequest{}).
		Where("id_request = ?", requestID).
		Updates(map[string]interface{}{
			"analysis_success": nil,
			"analysis_result":  nil,
			"confidence_score": nil,
		}).Error
}
