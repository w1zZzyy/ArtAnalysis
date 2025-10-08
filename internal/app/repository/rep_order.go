package repository

import (
	"errors"
	"fmt"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"gorm.io/gorm"
)

// GetDraftOrder возвращает черновик заявки для указанного пользователя (создателя).
func (r *Repository) GetDraftOrder(userID uint) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder

	err := r.db.Where("id_creator = ? AND order_status = ?", userID, model.StatusDraft).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CreateOrder создаёт новую заявку на анализ композиционного центра.
func (r *Repository) CreateOrder(order *model.AnalysisOrder) error {
	return r.db.Create(order).Error
}

// AddExpertToOrder добавляет эксперта в заявку (создает запись в таблице m-m).
func (r *Repository) AddExpertToOrder(orderID, expertID uint) error {
	// Проверяем, нет ли уже такого эксперта в заявке
	var count int64
	r.db.Model(&model.ExpertsToOrders{}).
		Where("id_order = ? AND id_artcenter = ?", orderID, expertID).
		Count(&count)
	if count > 0 {
		return errors.New("expert already linked to this order")
	}

	link := model.ExpertsToOrders{
		ID_order:     orderID,
		ID_artcenter: expertID,
	}
	return r.db.Create(&link).Error
}

// GetOrderWithExperts получает заявку со всеми экспертами и их связями.
func (r *Repository) GetOrderWithExperts(orderID uint) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder

	err := r.db.
		Preload("ExpertsLinks.ArtExpert"). // Загружаем через связующую таблицу
		First(&order, orderID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("error fetching order: %v", err)
	}

	// Проверяем, что заявка не удалена
	if order.OrderStatus == model.StatusDeleted {
		return nil, errors.New("order not found or has been deleted")
	}

	return &order, nil
}

// LogicallyDeleteOrder выполняет логическое удаление заявки через SQL UPDATE.
func (r *Repository) LogicallyDeleteOrder(orderID uint) error {
	result := r.db.Exec("UPDATE analysis_orders SET order_status = ? WHERE id_order = ?", model.StatusDeleted, orderID)
	return result.Error
}
