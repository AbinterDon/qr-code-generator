package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(qr *QRCodeHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{qr_token}", qr.Scan)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/qr_code", qr.Create)
		r.Get("/qr_code/{qr_token}", qr.Get)
		r.Put("/qr_code/{qr_token}", qr.Edit)
		r.Delete("/qr_code/{qr_token}", qr.Delete)
	})

	return r
}
