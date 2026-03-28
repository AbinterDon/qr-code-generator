//go:build integration

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/abinter/qr-code-generator/db"
	"github.com/abinter/qr-code-generator/internal/domain"
	"github.com/abinter/qr-code-generator/internal/repository"
)

func setupRepo(t *testing.T) *repository.PostgresRepository {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://qr:qr@localhost:5432/qrcode?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(context.Background(), "TRUNCATE TABLE qr_codes RESTART IDENTITY")
		pool.Close()
	})

	return repository.NewPostgresRepository(pool)
}

func TestPostgresRepository_CreateAndGet(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	qr := &domain.QRCode{
		UserID:  "user-1",
		QRToken: "tok123",
		URL:     "https://example.com",
	}

	if err := repo.Create(ctx, qr); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := repo.GetByToken(ctx, "tok123")
	if err != nil {
		t.Fatalf("GetByToken() error: %v", err)
	}
	if got.URL != qr.URL {
		t.Errorf("URL = %q; want %q", got.URL, qr.URL)
	}
	if got.UserID != qr.UserID {
		t.Errorf("UserID = %q; want %q", got.UserID, qr.UserID)
	}
}

func TestPostgresRepository_GetByToken_NotFound(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	_, err := repo.GetByToken(ctx, "nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("error = %v; want %v", err, domain.ErrNotFound)
	}
}

func TestPostgresRepository_Update(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	if err := repo.Create(ctx, &domain.QRCode{UserID: "u1", QRToken: "tok-upd", URL: "https://old.com"}); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := repo.Update(ctx, "tok-upd", "https://new.com"); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, _ := repo.GetByToken(ctx, "tok-upd")
	if got.URL != "https://new.com" {
		t.Errorf("URL after update = %q; want %q", got.URL, "https://new.com")
	}
}

func TestPostgresRepository_Delete(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	if err := repo.Create(ctx, &domain.QRCode{UserID: "u1", QRToken: "tok-del", URL: "https://example.com"}); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := repo.Delete(ctx, "tok-del"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := repo.GetByToken(ctx, "tok-del")
	if err != domain.ErrNotFound {
		t.Errorf("after delete: error = %v; want %v", err, domain.ErrNotFound)
	}
}

func TestPostgresRepository_GetByUserID(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	for i, tok := range []string{"tok-a", "tok-b", "tok-c"} {
		repo.Create(ctx, &domain.QRCode{UserID: "user-list", QRToken: tok, URL: "https://example.com"})
		_ = i
	}

	results, err := repo.GetByUserID(ctx, "user-list")
	if err != nil {
		t.Fatalf("GetByUserID() error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("count = %d; want 3", len(results))
	}
}

func TestPostgresRepository_Create_DuplicateToken(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	qr := &domain.QRCode{UserID: "u1", QRToken: "dup-tok", URL: "https://example.com"}
	if err := repo.Create(ctx, qr); err != nil {
		t.Fatalf("first Create() error: %v", err)
	}

	err := repo.Create(ctx, &domain.QRCode{UserID: "u2", QRToken: "dup-tok", URL: "https://other.com"})
	if err != domain.ErrTokenConflict {
		t.Errorf("duplicate error = %v; want %v", err, domain.ErrTokenConflict)
	}
}
