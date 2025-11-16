package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"gorm.io/gorm"
)

func (r *Repository) GetDraftOrder(userID uint) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder

	err := r.db.Where("id_creator = ? AND order_status = ?", userID, model.StatusDraft).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) CreateOrder(order *model.AnalysisOrder) error {
	if order.DateCreated.IsZero() {
		order.DateCreated = time.Now()
	}
	return r.db.Create(order).Error
}

func (r *Repository) AddExpertToOrder(orderID, expertID uint) error {
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

func (r *Repository) GetCurrentDraftOrder(userID uint) (*model.AnalysisOrder, int, error) {
	var order model.AnalysisOrder
	err := r.db.Preload("ExpertsLinks").Where("id_creator = ? AND order_status = ?", userID, "черновик").First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}
	count := len(order.ExpertsLinks)
	return &order, count, nil
}

func (r *Repository) ListAnalysisOrders(status, from, to string) ([]*model.AnalysisOrder, error) {
	var orders []*model.AnalysisOrder
	q := r.db.Preload("ExpertsLinks.ArtExpert").Preload("User")
	if status != "" {
		q = q.Where("order_status = ?", status)
	}
	if from != "" {
		q = q.Where("date_created >= ?", from)
	}
	if to != "" {
		q = q.Where("date_created <= ?", to)
	}
	if err := q.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *Repository) ListTaskByUser(userID uint, status, from, to string) ([]*model.AnalysisOrder, error) {
	var orders []*model.AnalysisOrder
	// Включим логирование SQL
	//r.db = r.db.Debug()
	query := r.db.Preload("GatesDegrees.Gate").Where("id_user = ?", userID)

	if status != "" {
		query = query.Where("order_status = ?", status)
	} else {
		query = query.Where("order_status NOT IN ?", []string{model.StatusDeleted, model.StatusDraft})
	}
	if from != "" {
		query = query.Where("date_created >= ?", from)
	}
	if to != "" {
		query = query.Where("date_created <= ?", to)
	}

	err := query.Order("date_created DESC").Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) GetAnalysisOrderByID(id uint) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder
	if err := r.db.Preload("ExpertsLinks.ArtExpert").Preload("User").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) UpdateAnalysisOrder(id uint, status *string, moderatorID *uint) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	if status != nil {
		order.OrderStatus = *status
	}
	if moderatorID != nil {
		order.ID_moderator = moderatorID
	}
	if err := r.db.Save(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) FormAnalysisOrder(id uint) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	if order.OrderStatus != "черновик" {
		return nil, errors.New("only draft orders can be formed")
	}
	order.OrderStatus = "сформирована"
	order.DateFormed = time.Now()
	if err := r.db.Save(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) ResolveAnalysisOrder(id uint, moderatorID uint, action string) (*model.AnalysisOrder, error) {
	var order model.AnalysisOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	if order.OrderStatus != "сформирована" {
		return nil, errors.New("only formed orders can be resolved")
	}
	if action == "complete" {
		order.OrderStatus = "завершена"
		order.DateCompleted = time.Now()
	} else if action == "reject" {
		order.OrderStatus = "отклонена"
		order.DateCompleted = time.Now()
	} else {
		return nil, errors.New("invalid action")
	}
	order.ID_moderator = &moderatorID
	if err := r.db.Save(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) DeleteAnalysisOrder(id uint) error {
	return r.db.Model(&model.AnalysisOrder{}).Where("id_order = ?", id).Update("order_status", "удалена").Error
}

func (r *Repository) RemoveExpertFromOrder(orderID, expertID uint) error {
	return r.db.Where("id_order = ? AND id_artcenter = ?", orderID, expertID).Delete(&model.ExpertsToOrders{}).Error
}

func (r *Repository) UpdateExpertInOrder(orderID, expertID uint, centerX, centerY *float32) error {
	var link model.ExpertsToOrders
	if err := r.db.First(&link, "id_order = ? AND id_artcenter = ?", orderID, expertID).Error; err != nil {
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
