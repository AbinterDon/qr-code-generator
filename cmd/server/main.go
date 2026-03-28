package main

import (
	"context"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"

	"github.com/abinter/qr-code-generator/db"
	"github.com/abinter/qr-code-generator/internal/config"
	"github.com/abinter/qr-code-generator/internal/domain"
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

	var repo domain.QRCodeRepository = repository.NewPostgresRepository(pool)

	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("redis url: %v", err)
		}
		client := redis.NewClient(opt)
		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatalf("redis ping: %v", err)
		}
		repo = repository.NewCachedRepository(repo, client, cfg.CacheTTL)
		log.Printf("redis cache enabled (ttl=%s)", cfg.CacheTTL)
	}

	uc := usecase.NewQRCodeUseCase(repo, cfg.BaseURL)
	qrHandler := handler.NewQRCodeHandler(uc)
	router := handler.NewRouter(qrHandler)

	log.Printf("server starting on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
