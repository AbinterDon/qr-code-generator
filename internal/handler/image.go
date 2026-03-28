package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/abinter/qr-code-generator/pkg/qrimage"
)

// GetImage handles GET /v1/qr_code_image/:qr_token
// Query params: dimension (int), color (hex), border (int)
func (h *QRCodeHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	qrToken := urlParam(r, "qr_token")

	qr, err := h.uc.GetQRCode(r.Context(), qrToken)
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	opts := qrimage.Options{
		// The QR code encodes the scan URL, not the raw destination URL.
		Content:   fmt.Sprintf("%s/%s", r.Host, qrToken),
		Dimension: queryInt(r, "dimension", 256),
		Color:     queryString(r, "color", "000000"),
		Border:    queryInt(r, "border", 4),
	}
	_ = qr // token is validated above; opts.Content uses the scan path

	png, err := qrimage.Generate(opts)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

// queryInt reads an integer query parameter, returning defaultVal on missing/invalid input.
func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

// queryString reads a string query parameter, returning defaultVal if empty.
func queryString(r *http.Request, key, defaultVal string) string {
	if s := r.URL.Query().Get(key); s != "" {
		return s
	}
	return defaultVal
}
