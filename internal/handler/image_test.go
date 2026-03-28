package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abinter/qr-code-generator/internal/domain"
	"github.com/abinter/qr-code-generator/internal/handler"
)

func TestGetImage_Success(t *testing.T) {
	uc := &mockUseCase{
		getFunc: func(ctx context.Context, qrToken string) (*domain.QRCode, error) {
			return &domain.QRCode{QRToken: qrToken, URL: "https://example.com"}, nil
		},
	}
	h := handler.NewQRCodeHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/v1/qr_code_image/tok123?dimension=256&color=000000&border=4", nil)
	req = handler.WithURLParam(req, "qr_token", "tok123")
	w := httptest.NewRecorder()

	h.GetImage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d (body: %s)", w.Code, http.StatusOK, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q; want image/png", ct)
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestGetImage_NotFound(t *testing.T) {
	uc := &mockUseCase{
		getFunc: func(ctx context.Context, qrToken string) (*domain.QRCode, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := handler.NewQRCodeHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/v1/qr_code_image/missing", nil)
	req = handler.WithURLParam(req, "qr_token", "missing")
	w := httptest.NewRecorder()

	h.GetImage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetImage_InvalidDimension(t *testing.T) {
	uc := &mockUseCase{
		getFunc: func(ctx context.Context, qrToken string) (*domain.QRCode, error) {
			return &domain.QRCode{QRToken: qrToken, URL: "https://example.com"}, nil
		},
	}
	h := handler.NewQRCodeHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/v1/qr_code_image/tok123?dimension=abc", nil)
	req = handler.WithURLParam(req, "qr_token", "tok123")
	w := httptest.NewRecorder()

	h.GetImage(w, req)

	// Invalid dimension falls back to default — should still succeed
	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}
}
