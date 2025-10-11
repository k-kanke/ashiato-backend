package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/k-kanke/ashiato-backend/pkg/api"
	"github.com/k-kanke/ashiato-backend/pkg/api/handler"
	"github.com/k-kanke/ashiato-backend/pkg/infra/database"
	"github.com/k-kanke/ashiato-backend/pkg/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// DB接続
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL not set in .env")
	}
	dbClient, err := database.NewDBClient(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer dbClient.DB.Close()

	// User関連
	userRepo := database.NewUserRepository(dbClient)
	userUc := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUc)

	// Pin関連
	pinRepo := database.NewPinRepository(dbClient)
	pinUc := usecase.NewPinUsecase(pinRepo)
	pinHandler := handler.NewPinHandler(pinUc)

	router := api.SetupRouter(userHandler, pinHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run :%v", err)
	}
}
