package main

import (
	"log"
	"net/http"

	"github.com/abinter/qr-code-generator/internal/config"
	"github.com/abinter/qr-code-generator/internal/handler"
	"github.com/abinter/qr-code-generator/internal/usecase"
)

func main() {
	cfg := config.Load()

	// TODO: replace with real DB repository in Stage 4
	uc := usecase.NewQRCodeUseCase(nil, cfg.BaseURL)

	qrHandler := handler.NewQRCodeHandler(uc)
	router := handler.NewRouter(qrHandler)

	log.Printf("server starting on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
