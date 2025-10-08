package main

import (
	"fmt"
	"os"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"
	"github.com/w1zZzyy22/art-analysis/internal/app/storage"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(storage.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&model.Users{},
		&model.ArtExpert{},
		&model.AnalysisOrder{},
		&model.ExpertsToOrders{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	sqlBytes, err := os.ReadFile("build/fill.sql")
	if err != nil {
		panic(err)
	}

	err = db.Exec(string(sqlBytes)).Error
	if err != nil {
		panic(err)
	}

	fmt.Println("Migration completed successfully")
}
