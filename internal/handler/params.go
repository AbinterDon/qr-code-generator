package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const paramKey contextKey = "urlParams"

// urlParam extracts a URL parameter from the request.
// Works with both chi router and test-injected params.
func urlParam(r *http.Request, key string) string {
	// Try chi first (production path)
	if val := chi.URLParam(r, key); val != "" {
		return val
	}
	// Fall back to test-injected params
	if params, ok := r.Context().Value(paramKey).(map[string]string); ok {
		return params[key]
	}
	return ""
}

// WithURLParam injects URL parameters into a request context for testing.
func WithURLParam(r *http.Request, key, value string) *http.Request {
	params, ok := r.Context().Value(paramKey).(map[string]string)
	if !ok {
		params = make(map[string]string)
	}
	params[key] = value
	ctx := context.WithValue(r.Context(), paramKey, params)
	return r.WithContext(ctx)
}
