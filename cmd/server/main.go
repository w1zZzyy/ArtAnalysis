package main

import (
	"log"

	"github.com/w1zZzyy22/art-analysis/internal/server"
	"github.com/w1zZzyy22/art-analysis/internal/storage"
)

func main() {
	storage.InitMinioClient()
	srv := server.NewServer()
	log.Println("Запуск сервера приложения 'Анализ композиционного центра на произведении искусства'...")
	srv.Start()
}
