package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/abinter/qr-code-generator/internal/domain"
	"github.com/abinter/qr-code-generator/internal/usecase"
)

func TestCreateQRCode(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		url       string
		repoErr   error
		wantErr   error
	}{
		{
			name:   "success",
			userID: "user-1",
			url:    "https://example.com",
		},
		{
			name:    "empty url",
			userID:  "user-1",
			url:     "",
			wantErr: domain.ErrInvalidURL,
		},
		{
			name:    "url too long",
			userID:  "user-1",
			url:     buildString(2049),
			wantErr: domain.ErrURLTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &domain.MockQRCodeRepository{
				CreateFunc: func(qrcode *domain.QRCode) error {
					return tt.repoErr
				},
			}
			uc := usecase.NewQRCodeUseCase(repo, "https://myqrcode.com")

			got, err := uc.CreateQRCode(context.Background(), tt.userID, tt.url)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateQRCode() error = %v; want %v", err, tt.wantErr)
			}
			if err == nil {
				if got.QRToken == "" {
					t.Error("CreateQRCode() returned empty QRToken")
				}
				if got.URL != tt.url {
					t.Errorf("CreateQRCode() URL = %q; want %q", got.URL, tt.url)
				}
				if got.UserID != tt.userID {
					t.Errorf("CreateQRCode() UserID = %q; want %q", got.UserID, tt.userID)
				}
			}
		})
	}
}

func TestGetQRCode(t *testing.T) {
	existing := &domain.QRCode{
		QRToken: "abc123",
		URL:     "https://example.com",
		UserID:  "user-1",
	}

	tests := []struct {
		name    string
		token   string
		repoRet *domain.QRCode
		repoErr error
		wantErr error
	}{
		{
			name:    "found",
			token:   "abc123",
			repoRet: existing,
		},
		{
			name:    "not found",
			token:   "missing",
			repoErr: domain.ErrNotFound,
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &domain.MockQRCodeRepository{
				GetByTokenFunc: func(token string) (*domain.QRCode, error) {
					return tt.repoRet, tt.repoErr
				},
			}
			uc := usecase.NewQRCodeUseCase(repo, "https://myqrcode.com")

			got, err := uc.GetQRCode(context.Background(), tt.token)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetQRCode() error = %v; want %v", err, tt.wantErr)
			}
			if err == nil && got.QRToken != tt.token {
				t.Errorf("GetQRCode() QRToken = %q; want %q", got.QRToken, tt.token)
			}
		})
	}
}

func TestEditQRCode(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		newURL  string
		repoErr error
		wantErr error
	}{
		{
			name:   "success",
			token:  "abc123",
			newURL: "https://new-url.com",
		},
		{
			name:    "empty url",
			token:   "abc123",
			newURL:  "",
			wantErr: domain.ErrInvalidURL,
		},
		{
			name:    "not found",
			token:   "missing",
			newURL:  "https://new-url.com",
			repoErr: domain.ErrNotFound,
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &domain.MockQRCodeRepository{
				UpdateFunc: func(token string, url string) error {
					return tt.repoErr
				},
			}
			uc := usecase.NewQRCodeUseCase(repo, "https://myqrcode.com")

			err := uc.EditQRCode(context.Background(), tt.token, tt.newURL)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("EditQRCode() error = %v; want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteQRCode(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		repoErr error
		wantErr error
	}{
		{
			name:  "success",
			token: "abc123",
		},
		{
			name:    "not found",
			token:   "missing",
			repoErr: domain.ErrNotFound,
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &domain.MockQRCodeRepository{
				DeleteFunc: func(token string) error {
					return tt.repoErr
				},
			}
			uc := usecase.NewQRCodeUseCase(repo, "https://myqrcode.com")

			err := uc.DeleteQRCode(context.Background(), tt.token)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteQRCode() error = %v; want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListQRCodes(t *testing.T) {
	repo := &domain.MockQRCodeRepository{
		GetByUserIDFunc: func(userID string) ([]*domain.QRCode, error) {
			return []*domain.QRCode{
				{QRToken: "tok1", UserID: userID},
				{QRToken: "tok2", UserID: userID},
			}, nil
		},
	}
	uc := usecase.NewQRCodeUseCase(repo, "https://myqrcode.com")

	got, err := uc.ListQRCodes(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ListQRCodes() unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListQRCodes() count = %d; want 2", len(got))
	}
}

func buildString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
