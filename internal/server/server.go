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

	s.mux.HandleFunc("/artcenters", handler.ArtCentersHandler)
	s.mux.HandleFunc("/artcenter/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/artcenter/")
		if id == "" {
			http.Error(w, "ID произведения не указан", http.StatusBadRequest)
			return
		}
		r = r.WithContext(r.Context())
		handler.ArtCenterDetailHandler(w, r.WithContext(r.Context()))
	})

	s.mux.HandleFunc("/basket/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/basket/")
		if id == "" {
			http.Error(w, "ID корзины не указан", http.StatusBadRequest)
			return
		}
		r = r.WithContext(r.Context())
		handler.BasketDetailHandler(w, r.WithContext(r.Context()))
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
