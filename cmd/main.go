package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kirillat6/urlShorter/internal/database"
	"github.com/kirillat6/urlShorter/internal/handler"
	repo "github.com/kirillat6/urlShorter/internal/repository"
	"github.com/kirillat6/urlShorter/internal/service"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Проблемы с загрузкой переменных окружений")
	}
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASSWORD")
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbName := os.Getenv("DATABASE_NAME")

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, connString)
	if err != nil {
		log.Fatalf("Проблемы с подключением БД: %v", err)
	}

	defer pool.Close()
	linkRepo := repo.NewLinkRepository(pool)
	linkService := service.NewLinkService(linkRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		handler.CreateLink(w, r, linkService)
	})
	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		handler.RedirectLink(w, r, linkService)
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Ошибка HTTP-сервера: %v", err)
	}

}
