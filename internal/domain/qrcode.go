package domain

import (
	"context"
	"time"
)

type QRCode struct {
	ID        int64
	UserID    string
	QRToken   string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type QRCodeRepository interface {
	Create(ctx context.Context, qrcode *QRCode) error
	GetByToken(ctx context.Context, token string) (*QRCode, error)
	GetByUserID(ctx context.Context, userID string) ([]*QRCode, error)
	Update(ctx context.Context, token string, url string) error
	Delete(ctx context.Context, token string) error
}
