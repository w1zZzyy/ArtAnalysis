package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/w1zZzyy22/art-analysis/internal/handler"
)

type Server struct {
	Port string
	mux  *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		Port: getPort(),
		mux:  http.NewServeMux(),
	}

	s.mux.HandleFunc("/services", handler.ServicesHandler)
	s.mux.HandleFunc("/service/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/service/")
		if id == "" {
			http.Error(w, "ID услуги не указан", http.StatusBadRequest)
			return
		}
		r = r.WithContext(r.Context())
		handler.ServiceDetailHandler(w, r.WithContext(r.Context()))
	})

	s.mux.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/order/")
		if id == "" {
			http.Error(w, "ID заявки не указан", http.StatusBadRequest)
			return
		}
		r = r.WithContext(r.Context())
		handler.OrderDetailHandler(w, r.WithContext(r.Context()))
	})

	fs := http.FileServer(http.Dir("./static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	return s
}

func (s *Server) Start() {
	addr := fmt.Sprintf(":%s", s.Port)
	log.Printf("Сервер запущен на http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, s.mux))
}

func getPort() string {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8080"
	}
	return p
}
