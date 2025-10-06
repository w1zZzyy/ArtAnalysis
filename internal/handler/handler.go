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
// /artcenters — список произведений
// ===============================
func ArtCentersHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	fmt.Println(q)

	artCenters := repository.GetAllArtCenters()
	filtered := []model.ArtCenter{}
	for _, s := range artCenters {
		if q == "" || strings.Contains(strings.ToLower(s.Title), strings.ToLower(q)) {
			filtered = append(filtered, s)
		}
	}

	basketID := "1"
	basketCount := 0
	basket, ok := repository.GetBasketByID(basketID)
	if ok {
		basketCount = len(strings.Split(basket.ArtIDs, ","))
	}

	type ArtCenterItem struct {
		ArtID          string
		Title          string
		Algorithm      string
		ArtDescription string
		ArtImageURL    string
	}

	var artCenterItems []ArtCenterItem
	for _, s := range filtered {
		artCenterItems = append(artCenterItems, ArtCenterItem{
			ArtID:          s.ArtID,
			Title:          s.Title,
			Algorithm:      s.Algorithm,
			ArtDescription: s.ArtDescription,
			ArtImageURL:    storage.BuildImageURL(s.ArtImageKey),
		})
	}

	tmplData := struct {
		ArtCenters  []ArtCenterItem
		Query       string
		BasketID    string
		BasketCount int
	}{
		ArtCenters:  artCenterItems,
		Query:       q,
		BasketID:    basketID,
		BasketCount: basketCount,
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
func ArtCenterDetailHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Не указан ID услуги", http.StatusBadRequest)
		return
	}
	id := parts[2]

	artCenter, ok := repository.GetArtCenterByID(id)
	if !ok {
		http.Error(w, "Произведение не найдено", http.StatusNotFound)
		return
	}

	data := struct {
		ArtCenter   model.ArtCenter
		ArtImageURL string
	}{
		ArtCenter:   *artCenter,
		ArtImageURL: storage.BuildImageURL(artCenter.ArtImageKey),
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
func BasketDetailHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Не указан ID заявки", http.StatusBadRequest)
		return
	}
	id := parts[2]

	basket, ok := repository.GetBasketByID(id)
	if !ok {
		http.Error(w, "Корзина не найдена", http.StatusNotFound)
		return
	}

	type BasketItem struct {
		ArtCenter   model.ArtCenter
		ArtImageURL string
		CenterX     int
		CenterY     int
	}

	var items []BasketItem
	var sumX, sumY float64
	artIDs := strings.Split(basket.ArtIDs, ",")

	for _, aid := range artIDs {
		artCenter, ok := repository.GetArtCenterByID(aid)
		if !ok {
			continue
		}

		result, ok := repository.GetAnalysisResultByBasketID(&basket.BasketID, &aid)
		if !ok {
			continue
		}
		coords := strings.Split(*result, ",")
		if len(coords) != 2 {
			continue
		}

		x, err1 := strconv.Atoi(coords[0])
		y, err2 := strconv.Atoi(coords[1])

		if err1 != nil || err2 != nil {
			continue
		}

		sumX += float64(x)
		sumY += float64(y)

		items = append(items, BasketItem{
			ArtCenter:   *artCenter,
			ArtImageURL: storage.BuildImageURL(artCenter.ArtImageKey),
			CenterX:     x,
			CenterY:     y,
		})
	}

	cnt := float64(len(items))
	globalResult := ""

	if cnt > 0 {
		globalResult = fmt.Sprintf("X: %.2f\t Y: %.2f", sumX/cnt, sumY/cnt)
	}

	data := struct {
		Basket       model.Basket
		Items        []BasketItem
		GlobalResult string
	}{
		Basket:       *basket,
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
