package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abinter/qr-code-generator/internal/domain"
	"github.com/abinter/qr-code-generator/internal/handler"
)

// mockUseCase implements handler.QRCodeUseCaseInterface
type mockUseCase struct {
	createFunc func(ctx context.Context, userID, url string) (*domain.QRCode, error)
	getFunc    func(ctx context.Context, qrToken string) (*domain.QRCode, error)
	editFunc   func(ctx context.Context, qrToken, newURL string) error
	deleteFunc func(ctx context.Context, qrToken string) error
	listFunc   func(ctx context.Context, userID string) ([]*domain.QRCode, error)
}

func (m *mockUseCase) CreateQRCode(ctx context.Context, userID, url string) (*domain.QRCode, error) {
	return m.createFunc(ctx, userID, url)
}
func (m *mockUseCase) GetQRCode(ctx context.Context, qrToken string) (*domain.QRCode, error) {
	return m.getFunc(ctx, qrToken)
}
func (m *mockUseCase) EditQRCode(ctx context.Context, qrToken, newURL string) error {
	return m.editFunc(ctx, qrToken, newURL)
}
func (m *mockUseCase) DeleteQRCode(ctx context.Context, qrToken string) error {
	return m.deleteFunc(ctx, qrToken)
}
func (m *mockUseCase) ListQRCodes(ctx context.Context, userID string) ([]*domain.QRCode, error) {
	return m.listFunc(ctx, userID)
}

func TestCreateQRCode(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		ucErr      error
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"url":"https://example.com"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid url",
			body:       `{"url":""}`,
			ucErr:      domain.ErrInvalidURL,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "internal error",
			body:       `{"url":"https://example.com"}`,
			ucErr:      errors.New("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUseCase{
				createFunc: func(ctx context.Context, userID, url string) (*domain.QRCode, error) {
					if tt.ucErr != nil {
						return nil, tt.ucErr
					}
					return &domain.QRCode{QRToken: "tok123", URL: url}, nil
				},
			}
			h := handler.NewQRCodeHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/v1/qr_code", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestGetQRCode(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		ucRet      *domain.QRCode
		ucErr      error
		wantStatus int
	}{
		{
			name:       "found",
			token:      "tok123",
			ucRet:      &domain.QRCode{QRToken: "tok123", URL: "https://example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			token:      "missing",
			ucErr:      domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUseCase{
				getFunc: func(ctx context.Context, qrToken string) (*domain.QRCode, error) {
					return tt.ucRet, tt.ucErr
				},
			}
			h := handler.NewQRCodeHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/v1/qr_code/"+tt.token, nil)
			req = handler.WithURLParam(req, "qr_token", tt.token)
			w := httptest.NewRecorder()

			h.Get(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		ucRet      *domain.QRCode
		ucErr      error
		wantStatus int
	}{
		{
			name:       "redirect",
			token:      "tok123",
			ucRet:      &domain.QRCode{QRToken: "tok123", URL: "https://example.com"},
			wantStatus: http.StatusFound,
		},
		{
			name:       "not found",
			token:      "missing",
			ucErr:      domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUseCase{
				getFunc: func(ctx context.Context, qrToken string) (*domain.QRCode, error) {
					return tt.ucRet, tt.ucErr
				},
			}
			h := handler.NewQRCodeHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.token, nil)
			req = handler.WithURLParam(req, "qr_token", tt.token)
			w := httptest.NewRecorder()

			h.Scan(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusFound {
				if loc := w.Header().Get("Location"); loc != tt.ucRet.URL {
					t.Errorf("Location = %q; want %q", loc, tt.ucRet.URL)
				}
			}
		})
	}
}

func TestEditQRCode(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		body       string
		ucErr      error
		wantStatus int
	}{
		{
			name:       "success",
			token:      "tok123",
			body:       `{"url":"https://new.com"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "not found",
			token:      "missing",
			body:       `{"url":"https://new.com"}`,
			ucErr:      domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUseCase{
				editFunc: func(ctx context.Context, qrToken, newURL string) error {
					return tt.ucErr
				},
			}
			h := handler.NewQRCodeHandler(uc)

			req := httptest.NewRequest(http.MethodPut, "/v1/qr_code/"+tt.token, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = handler.WithURLParam(req, "qr_token", tt.token)
			w := httptest.NewRecorder()

			h.Edit(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestDeleteQRCode(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		ucErr      error
		wantStatus int
	}{
		{
			name:       "success",
			token:      "tok123",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "not found",
			token:      "missing",
			ucErr:      domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUseCase{
				deleteFunc: func(ctx context.Context, qrToken string) error {
					return tt.ucErr
				},
			}
			h := handler.NewQRCodeHandler(uc)

			req := httptest.NewRequest(http.MethodDelete, "/v1/qr_code/"+tt.token, nil)
			req = handler.WithURLParam(req, "qr_token", tt.token)
			w := httptest.NewRecorder()

			h.Delete(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCreateQRCode_ResponseBody(t *testing.T) {
	uc := &mockUseCase{
		createFunc: func(ctx context.Context, userID, url string) (*domain.QRCode, error) {
			return &domain.QRCode{QRToken: "tok123", URL: url}, nil
		},
	}
	h := handler.NewQRCodeHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/v1/qr_code", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["qr_token"] != "tok123" {
		t.Errorf("qr_token = %q; want %q", resp["qr_token"], "tok123")
	}
}
