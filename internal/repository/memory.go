package repository

import (
	"strconv"
	"strings"

	"github.com/w1zZzyy22/art-analysis/internal/model"
)

// ===============================
// Коллекции в памяти
// ===============================

// ArtCenters — список произведений по анализу композиционного центра
var ArtCenters = []model.ArtCenter{
	{
		ArtID:          "1",
		Title:          "Анализ композиционного центра картины",
		Algorithm:      "Визуальный анализ изображения",
		ArtDescription: "Определение ключевой точки композиции, выявление фокуса и направления взгляда.",
		ArtImageKey:    "abstract_1.jpg",
	},
	{
		ArtID:          "2",
		Title:          "Цветовой анализ произведений",
		Algorithm:      "Анализ цветовой гармонии изображения",
		ArtDescription: "Определение доминирующих цветов, контрастов и гармонии в картине или фотографии.",
		ArtImageKey:    "abstract_2.jpg",
	},
	{
		ArtID:          "3",
		Title:          "Оценка композиции фотографий",
		Algorithm:      "Цифровой анализ",
		ArtDescription: "Выявление сильных и слабых сторон композиции фотографии, рекомендации по улучшению.",
		ArtImageKey:    "abstract_3.jpg",
	},
	{
		ArtID:          "4",
		Title:          "Анализ скульптур и объектов",
		Algorithm:      "3D визуализация",
		ArtDescription: "Определение композиционного центра и перспективного восприятия объема объекта.",
		ArtImageKey:    "abstract_4.jpg",
	},
	{
		ArtID:          "5",
		Title:          "Композиционный анализ иллюстраций",
		Algorithm:      "Визуальный и цифровой анализ",
		ArtDescription: "Определение ключевых элементов иллюстрации и построение визуального фокуса.",
		ArtImageKey:    "abstract_5.jpg",
	},
}

// Baskets — список корзин
var Baskets = []model.Basket{
	{
		BasketID: "1",
		ArtIDs:   "1,3,5", // IDs произведений из ArtCenters через запятую
	},
}

var Results = []model.AnalysisResult{
	{
		BasketID: "1",
		Results:  map[string]string{"1": "10,20", "3": "24,92", "5": "88,11"},
	},
}

// ===============================
// Функции доступа
// ===============================

// GetAllArtCenters возвращает все произведения
func GetAllArtCenters() []model.ArtCenter {
	return ArtCenters
}

// GetArtCenterByID возвращает произведение по ID
func GetArtCenterByID(id string) (*model.ArtCenter, bool) {
	for _, s := range ArtCenters {
		if s.ArtID == id {
			return &s, true
		}
	}
	return nil, false
}

// GetBasketByID возвращает корзину по ID
func GetBasketByID(id string) (*model.Basket, bool) {
	for _, o := range Baskets {
		if o.BasketID == id {
			return &o, true
		}
	}
	return nil, false
}

func GetAnalysisResultByBasketID(basketID *string, artId *string) (*string, bool) {
	for _, r := range Results {
		if r.BasketID == *basketID {
			if coords, ok := r.Results[*artId]; ok {
				return &coords, true
			}
			return nil, false
		}
	}
	return nil, false
}

// CountItemsInBasket возвращает сумму всех количеств произведений в корзине
func CountItemsInBasket(basket *model.Basket) int {
	counts := strings.Split(basket.Counts, ",")
	total := 0
	for _, c := range counts {
		n, err := strconv.Atoi(c)
		if err != nil {
			continue
		}
		total += n
	}
	return total
}

// ===============================
// Примечание по изображению и Minio
// ===============================
// В поле ImageKey хранится имя файла изображения, например "service1.jpg".
// URL для доступа к изображению можно формировать так:
//   https://<MINIO_ENDPOINT>/<BUCKET_NAME>/<ImageKey>
// где <MINIO_ENDPOINT> и <BUCKET_NAME> берутся из переменных окружения.
// Для локального тестирования можно хранить изображения в static/images/.
