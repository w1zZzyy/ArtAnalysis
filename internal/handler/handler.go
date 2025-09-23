package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/w1zZzyy22/art-analysis/internal/model"
	"github.com/w1zZzyy22/art-analysis/internal/repository"
	"github.com/w1zZzyy22/art-analysis/internal/storage"
)

// ===============================
// /services — список услуг
// ===============================
func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	services := repository.GetAllServices()
	filtered := []model.Service{}
	for _, s := range services {
		if q == "" || strings.Contains(strings.ToLower(s.Name), strings.ToLower(q)) {
			filtered = append(filtered, s)
		}
	}

	orderID := "1"
	orderCount := 0
	order, ok := repository.GetOrderByID(orderID)
	if ok {
		orderCount = len(strings.Split(order.ItemIDs, ","))
	}

	type ServiceItem struct {
		ID       string
		Name     string
		Method   string
		ImageURL string
	}

	var serviceItems []ServiceItem
	for _, s := range filtered {
		serviceItems = append(serviceItems, ServiceItem{
			ID:       s.ID,
			Name:     s.Name,
			Method:   s.Method,
			ImageURL: storage.BuildImageURL(s.ImageKey),
		})
	}

	tmplData := struct {
		Services   []ServiceItem
		Query      string
		OrderID    string
		OrderCount int
	}{
		Services:   serviceItems,
		Query:      q,
		OrderID:    orderID,
		OrderCount: orderCount,
	}

	tmpl, err := template.ParseFiles("templates/services.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона services.html: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, tmplData); err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ===============================
// /service/{id} — детальная страница услуги
// ===============================
func ServiceDetailHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Не указан ID услуги", http.StatusBadRequest)
		return
	}
	id := parts[2]

	service, ok := repository.GetServiceByID(id)
	if !ok {
		http.Error(w, "Услуга не найдена", http.StatusNotFound)
		return
	}

	data := struct {
		Service  model.Service
		ImageURL string
	}{
		Service:  *service,
		ImageURL: storage.BuildImageURL(service.ImageKey),
	}

	tmpl, err := template.ParseFiles("templates/service.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона service.html: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ===============================
// /order/{id} — страница заявки
// ===============================
func OrderDetailHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Не указан ID заявки", http.StatusBadRequest)
		return
	}
	id := parts[2]

	order, ok := repository.GetOrderByID(id)
	if !ok {
		http.Error(w, "Заявка не найдена", http.StatusNotFound)
		return
	}

	itemIDs := strings.Split(order.ItemIDs, ",")
	counts := strings.Split(order.Counts, ",")

	type OrderItem struct {
		Service  model.Service
		Quantity int
		ImageURL string
		Result   string
	}

	var items []OrderItem
	var sumX, sumY float64
	var count float64

	for i, sid := range itemIDs {
		service, ok := repository.GetServiceByID(sid)
		if !ok {
			continue
		}

		qty := 1
		if i < len(counts) {
			if n, err := strconv.Atoi(counts[i]); err == nil {
				qty = n
			}
		}

		result := ""
		if i < len(order.Results) {
			result = order.Results[i]
			coords := strings.Split(result, ",")
			if len(coords) == 2 {
				if x, err1 := strconv.ParseFloat(coords[0], 64); err1 == nil {
					if y, err2 := strconv.ParseFloat(coords[1], 64); err2 == nil {
						sumX += x
						sumY += y
						count++
					}
				}
			}
		}

		items = append(items, OrderItem{
			Service:  *service,
			Quantity: qty,
			ImageURL: storage.BuildImageURL(service.ImageKey),
			Result:   result,
		})
	}

	globalResult := ""
	if count > 0 {
		globalResult = fmt.Sprintf("x: %.2f, y: %.2f", sumX/count, sumY/count)
	}

	data := struct {
		Order        model.Order
		Items        []OrderItem
		GlobalResult string
	}{
		Order:        *order,
		Items:        items,
		GlobalResult: globalResult,
	}

	tmpl, err := template.ParseFiles("templates/order.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона order.html: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
