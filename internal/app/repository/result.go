package repository

import (
	"fmt"
	"math"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"
	"gorm.io/gorm"
)

// CalculateArtAnalysis вычисляет итоговый композиционный центр на основе данных экспертов
func (r *Repository) CalculateArtAnalysis(requestID uint) error {
	var request model.CenterRequest

	// Загружаем заявку со всеми связанными экспертами и их центрами
	err := r.db.
		Preload("ExpertsLinks", func(db *gorm.DB) *gorm.DB {
			// Берем только экспертов, которые ввели свои координаты
			return db.Where("center_x IS NOT NULL AND center_y IS NOT NULL")
		}).
		Preload("ExpertsLinks.ArtExpert").
		First(&request, requestID).Error

	if err != nil {
		return fmt.Errorf("ошибка загрузки заявки: %v", err)
	}

	// Если нет экспертов с координатами
	if len(request.ExpertsLinks) == 0 {
		// Ставим центр в середину (0.5, 0.5)
		centerX := float32(0.5)
		centerY := float32(0.5)

		return r.db.Model(&request).Updates(map[string]interface{}{
			"factor_x": centerX,
			"factor_y": centerY,
		}).Error
	}

	// 1. Считаем простое среднее арифметическое всех координат
	var sumX, sumY float64
	for _, expert := range request.ExpertsLinks {
		if expert.CenterX != nil && expert.CenterY != nil {
			sumX += float64(*expert.CenterX)
			sumY += float64(*expert.CenterY)
		}
	}

	// Средние значения
	avgX := float32(sumX / float64(len(request.ExpertsLinks)))
	avgY := float32(sumY / float64(len(request.ExpertsLinks)))

	// 2. Корректируем в зависимости от алгоритмов экспертов
	var correctedX, correctedY float64
	var totalWeight float64

	for _, expert := range request.ExpertsLinks {
		if expert.CenterX != nil && expert.CenterY != nil {
			// Вес алгоритма (сколько мы ему доверяем)
			weight := r.getAlgorithmWeight(expert.ArtExpert.Algorithm)

			// Берем координаты эксперта
			x := float64(*expert.CenterX)
			y := float64(*expert.CenterY)

			// Корректируем в зависимости от типа алгоритма
			switch expert.ArtExpert.Algorithm {
			case "Цифровой анализ":
				// Цифровой анализ точный, но может быть смещен к центру
				x = x*0.9 + 0.5*0.1
				y = y*0.9 + 0.5*0.1
			case "3D визуализация":
				// 3D анализ часто смещен для создания глубины
				y = y * 0.85
			case "Визуальный анализ изображения":
				// Художественный анализ стремится к золотому сечению
				x = x*0.8 + 0.618*0.2
				y = y*0.8 + 0.618*0.2
			}

			correctedX += x * float64(weight)
			correctedY += y * float64(weight)
			totalWeight += float64(weight)
		}
	}

	// Взвешенное среднее
	weightedX := float32(correctedX / totalWeight)
	weightedY := float32(correctedY / totalWeight)

	// 3. Финализируем результат - смесь среднего и взвешенного
	// Если эксперты согласны (разброс маленький) - больше доверяем взвешенному
	// Если разброс большой - больше доверяем простому среднему

	// Считаем разброс
	var spread float64
	for _, expert := range request.ExpertsLinks {
		if expert.CenterX != nil && expert.CenterY != nil {
			dx := float64(*expert.CenterX) - float64(avgX)
			dy := float64(*expert.CenterY) - float64(avgY)
			spread += math.Sqrt(dx*dx + dy*dy)
		}
	}
	spread /= float64(len(request.ExpertsLinks))

	// Коэффициент доверия взвешенному среднему
	trustWeighted := 1.0 / (1.0 + spread)

	// Итоговые координаты
	finalX := weightedX*float32(trustWeighted) + avgX*float32(1-trustWeighted)
	finalY := weightedY*float32(trustWeighted) + avgY*float32(1-trustWeighted)

	// 4. Гарантируем диапазон 0-1
	if finalX < 0 {
		finalX = 0
	} else if finalX > 1 {
		finalX = 1
	}
	if finalY < 0 {
		finalY = 0
	} else if finalY > 1 {
		finalY = 1
	}

	// 5. Сохраняем
	return r.db.Model(&request).Updates(map[string]interface{}{
		"factor_x": finalX,
		"factor_y": finalY,
	}).Error
}

// getAlgorithmWeight возвращает вес доверия алгоритму
func (r *Repository) getAlgorithmWeight(algorithm string) float32 {
	switch algorithm {
	case "Цифровой анализ":
		return 1.2 // Самый точный
	case "Визуальный анализ изображения":
		return 1.0
	case "Визуальный и цифровой анализ":
		return 1.1
	case "Анализ цветовой гармонии изображения":
		return 0.9
	case "3D визуализация":
		return 0.8
	default:
		return 1.0
	}
}
