// @title Art Analysis API
// @version 1.0
// @description API для анализа композиционного центра произведений искусства
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: "Bearer {token}"
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Импортируем сгенерированную документацию
	_ "github.com/w1zZzyy22/art-analysis/docs"

	"github.com/w1zZzyy22/art-analysis/internal/app/config"
	"github.com/w1zZzyy22/art-analysis/internal/app/handler"
	"github.com/w1zZzyy22/art-analysis/internal/app/redis"
	"github.com/w1zZzyy22/art-analysis/internal/app/repository"
	"github.com/w1zZzyy22/art-analysis/internal/app/storage"
	"github.com/w1zZzyy22/art-analysis/internal/pkg"
)

func main() {
	router := gin.Default()

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health-check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Art Analysis API is running",
		})
	})

	// Загружаем конфиг
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := storage.FromEnv()
	fmt.Println(postgresString)

	// Подключаем репозиторий
	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Подключаем Redis
	redisClient, errRedis := redis.New(context.Background(), conf.Redis)
	if errRedis != nil {
		logrus.Fatalf("error initializing redis: %v", errRedis)
	}

	hand := handler.NewHandler(rep, redisClient, &conf.JWT)

	application := pkg.NewApp(conf, router, hand)

	// Информационные сообщения
	fmt.Println("=== Art Analysis API ===")
	fmt.Println("Server started on: http://localhost:8080")
	fmt.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	fmt.Println("Health check: http://localhost:8080/health")
	fmt.Println("=========================")

	application.RunApp()
}
