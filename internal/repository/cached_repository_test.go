package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/abinter/qr-code-generator/internal/domain"
	"github.com/abinter/qr-code-generator/internal/repository"
)

func setupCachedRepo(t *testing.T) (*repository.CachedRepository, *domain.MockQRCodeRepository, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mock := &domain.MockQRCodeRepository{}
	cached := repository.NewCachedRepository(mock, client, time.Minute)
	return cached, mock, mr
}

func TestCachedRepository_GetByToken_CacheMiss_ThenHit(t *testing.T) {
	cached, mock, _ := setupCachedRepo(t)
	ctx := context.Background()

	callCount := 0
	mock.GetByTokenFunc = func(ctx context.Context, token string) (*domain.QRCode, error) {
		callCount++
		return &domain.QRCode{QRToken: token, URL: "https://example.com"}, nil
	}

	// First call: cache miss → hits DB
	qr1, err := cached.GetByToken(ctx, "tok123")
	if err != nil {
		t.Fatalf("first GetByToken() error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 DB call, got %d", callCount)
	}

	// Second call: cache hit → skips DB
	qr2, err := cached.GetByToken(ctx, "tok123")
	if err != nil {
		t.Fatalf("second GetByToken() error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 DB call after cache hit, got %d", callCount)
	}
	if qr1.URL != qr2.URL {
		t.Errorf("cached URL %q != original %q", qr2.URL, qr1.URL)
	}
}

func TestCachedRepository_GetByToken_NotFound_NotCached(t *testing.T) {
	cached, mock, _ := setupCachedRepo(t)
	ctx := context.Background()

	callCount := 0
	mock.GetByTokenFunc = func(ctx context.Context, token string) (*domain.QRCode, error) {
		callCount++
		return nil, domain.ErrNotFound
	}

	_, err := cached.GetByToken(ctx, "missing")
	if err != domain.ErrNotFound {
		t.Errorf("error = %v; want ErrNotFound", err)
	}

	// Not-found results must NOT be cached; DB should be called again
	_, _ = cached.GetByToken(ctx, "missing")
	if callCount != 2 {
		t.Errorf("expected 2 DB calls (not-found not cached), got %d", callCount)
	}
}

func TestCachedRepository_Update_InvalidatesCache(t *testing.T) {
	cached, mock, _ := setupCachedRepo(t)
	ctx := context.Background()

	callCount := 0
	mock.GetByTokenFunc = func(ctx context.Context, token string) (*domain.QRCode, error) {
		callCount++
		return &domain.QRCode{QRToken: token, URL: "https://old.com"}, nil
	}
	mock.UpdateFunc = func(ctx context.Context, token, url string) error { return nil }

	// Populate cache
	cached.GetByToken(ctx, "tok123")

	// Update should invalidate
	if err := cached.Update(ctx, "tok123", "https://new.com"); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	// Next get must hit DB again
	cached.GetByToken(ctx, "tok123")
	if callCount != 2 {
		t.Errorf("expected 2 DB calls after cache invalidation, got %d", callCount)
	}
}

func TestCachedRepository_Delete_InvalidatesCache(t *testing.T) {
	cached, mock, _ := setupCachedRepo(t)
	ctx := context.Background()

	callCount := 0
	mock.GetByTokenFunc = func(ctx context.Context, token string) (*domain.QRCode, error) {
		callCount++
		return &domain.QRCode{QRToken: token, URL: "https://example.com"}, nil
	}
	mock.DeleteFunc = func(ctx context.Context, token string) error { return nil }

	// Populate cache
	cached.GetByToken(ctx, "tok123")

	// Delete should invalidate
	if err := cached.Delete(ctx, "tok123"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	// Next get must hit DB again
	cached.GetByToken(ctx, "tok123")
	if callCount != 2 {
		t.Errorf("expected 2 DB calls after delete invalidation, got %d", callCount)
	}
}

func TestCachedRepository_CacheTTL_Expiry(t *testing.T) {
	cached, mock, mr := setupCachedRepo(t)
	ctx := context.Background()

	callCount := 0
	mock.GetByTokenFunc = func(ctx context.Context, token string) (*domain.QRCode, error) {
		callCount++
		return &domain.QRCode{QRToken: token, URL: "https://example.com"}, nil
	}

	cached.GetByToken(ctx, "tok123")

	// Fast-forward miniredis clock past TTL
	mr.FastForward(2 * time.Minute)

	cached.GetByToken(ctx, "tok123")
	if callCount != 2 {
		t.Errorf("expected 2 DB calls after TTL expiry, got %d", callCount)
	}
}

func TestCachedRepository_Create_PassesThrough(t *testing.T) {
	cached, mock, _ := setupCachedRepo(t)
	ctx := context.Background()

	called := false
	mock.CreateFunc = func(ctx context.Context, qr *domain.QRCode) error {
		called = true
		return nil
	}

	err := cached.Create(ctx, &domain.QRCode{QRToken: "tok", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if !called {
		t.Error("expected underlying repo Create to be called")
	}
}
