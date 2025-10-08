package main

import (
	"fmt"
	"log"

	"github.com/w1zZzyy22/art-analysis/internal/app/config"
	"github.com/w1zZzyy22/art-analysis/internal/app/handler"
	"github.com/w1zZzyy22/art-analysis/internal/app/repository"
	"github.com/w1zZzyy22/art-analysis/internal/app/storage"
	"github.com/w1zZzyy22/art-analysis/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := storage.FromEnv()
	fmt.Println("Postgres connection string:", postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	log.Println("Запуск сервера приложения 'Анализ композиционного центра на произведении искусства'...")
	application.RunApp()
}
