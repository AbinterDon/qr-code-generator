package main

import (
	"context"
	"log"
	"net/http"

	"github.com/abinter/qr-code-generator/db"
	"github.com/abinter/qr-code-generator/internal/config"
	"github.com/abinter/qr-code-generator/internal/handler"
	"github.com/abinter/qr-code-generator/internal/repository"
	"github.com/abinter/qr-code-generator/internal/usecase"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)
	uc := usecase.NewQRCodeUseCase(repo, cfg.BaseURL)

	qrHandler := handler.NewQRCodeHandler(uc)
	router := handler.NewRouter(qrHandler)

	log.Printf("server starting on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
