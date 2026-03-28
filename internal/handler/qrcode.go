package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/abinter/qr-code-generator/internal/domain"
)

// QRCodeUseCaseInterface defines the use case methods the handler depends on.
type QRCodeUseCaseInterface interface {
	CreateQRCode(ctx context.Context, userID, url string) (*domain.QRCode, error)
	GetQRCode(ctx context.Context, qrToken string) (*domain.QRCode, error)
	EditQRCode(ctx context.Context, qrToken, newURL string) error
	DeleteQRCode(ctx context.Context, qrToken string) error
	ListQRCodes(ctx context.Context, userID string) ([]*domain.QRCode, error)
}

type QRCodeHandler struct {
	uc QRCodeUseCaseInterface
}

func NewQRCodeHandler(uc QRCodeUseCaseInterface) *QRCodeHandler {
	return &QRCodeHandler{uc: uc}
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	QRToken string `json:"qr_token"`
}

type getResponse struct {
	URL string `json:"url"`
}

// Create handles POST /v1/qr_code
func (h *QRCodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: extract real user ID from auth middleware
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	qr, err := h.uc.CreateQRCode(r.Context(), userID, req.URL)
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createResponse{QRToken: qr.QRToken})
}

// Get handles GET /v1/qr_code/:qr_token
func (h *QRCodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	qrToken := urlParam(r, "qr_token")

	qr, err := h.uc.GetQRCode(r.Context(), qrToken)
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, getResponse{URL: qr.URL})
}

// Scan handles GET /:qr_token — 302 redirect to original URL
func (h *QRCodeHandler) Scan(w http.ResponseWriter, r *http.Request) {
	qrToken := urlParam(r, "qr_token")

	qr, err := h.uc.GetQRCode(r.Context(), qrToken)
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	http.Redirect(w, r, qr.URL, http.StatusFound)
}

// Edit handles PUT /v1/qr_code/:qr_token
func (h *QRCodeHandler) Edit(w http.ResponseWriter, r *http.Request) {
	qrToken := urlParam(r, "qr_token")

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.EditQRCode(r.Context(), qrToken, req.URL); err != nil {
		writeUseCaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /v1/qr_code/:qr_token
func (h *QRCodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	qrToken := urlParam(r, "qr_token")

	if err := h.uc.DeleteQRCode(r.Context(), qrToken); err != nil {
		writeUseCaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// helpers

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidURL), errors.Is(err, domain.ErrURLTooLong):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
