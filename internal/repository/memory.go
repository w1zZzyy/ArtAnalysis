package repository

import (
	"strconv"
	"strings"

	"github.com/w1zZzyy22/art-analysis/internal/model"
)

// ===============================
// Коллекции в памяти
// ===============================

// Services — список услуг по анализу композиционного центра
var Services = []model.Service{
	{
		ID:          "1",
		Name:        "Анализ композиционного центра картины",
		Method:      "Визуальный анализ изображения",
		Description: "Определение ключевой точки композиции, выявление фокуса и направления взгляда.",
		ImageKey:    "abstract_1.jpg",
	},
	{
		ID:          "2",
		Name:        "Цветовой анализ произведений",
		Method:      "Анализ цветовой гармонии изображения",
		Description: "Определение доминирующих цветов, контрастов и гармонии в картине или фотографии.",
		ImageKey:    "abstract_2.jpg",
	},
	{
		ID:          "3",
		Name:        "Оценка композиции фотографий",
		Method:      "Цифровой анализ",
		Description: "Выявление сильных и слабых сторон композиции фотографии, рекомендации по улучшению.",
		ImageKey:    "abstract_3.jpg",
	},
	{
		ID:          "4",
		Name:        "Анализ скульптур и объектов",
		Method:      "3D визуализация",
		Description: "Определение композиционного центра и перспективного восприятия объема объекта.",
		ImageKey:    "abstract_4.jpg",
	},
	{
		ID:          "5",
		Name:        "Композиционный анализ иллюстраций",
		Method:      "Визуальный и цифровой анализ",
		Description: "Определение ключевых элементов иллюстрации и построение визуального фокуса.",
		ImageKey:    "abstract_5.jpg",
	},
	{
		ID:          "6",
		Name:        "Анализ перспективы в архитектуре",
		Method:      "Цифровой и линейный анализ",
		Description: "Определение центров перспективы и композиционного построения в архитектурных объектах.",
		ImageKey:    "abstract_6.jpg",
	},
}

// Orders — список заявок
var Orders = []model.Order{
	{
		ID:      "1",
		ItemIDs: "1,3,5", // IDs услуг из Services через запятую
		Counts:  "2,1,1", // количество каждой услуги через запятую
		Results: []string{"10,20", "24,92", "88,11"},
	},
}

// ===============================
// Функции доступа
// ===============================

// GetAllServices возвращает все услуги
func GetAllServices() []model.Service {
	return Services
}

// GetServiceByID возвращает услугу по ID
func GetServiceByID(id string) (*model.Service, bool) {
	for _, s := range Services {
		if s.ID == id {
			return &s, true
		}
	}
	return nil, false
}

// GetOrderByID возвращает заявку по ID
func GetOrderByID(id string) (*model.Order, bool) {
	for _, o := range Orders {
		if o.ID == id {
			return &o, true
		}
	}
	return nil, false
}

// CountItemsInOrder возвращает сумму всех количеств услуг в заявке
func CountItemsInOrder(order *model.Order) int {
	counts := strings.Split(order.Counts, ",")
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
