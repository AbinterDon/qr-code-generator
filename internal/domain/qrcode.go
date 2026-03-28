package domain

import "time"

type QRCode struct {
	ID        int64
	UserID    string
	QRToken   string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type QRCodeRepository interface {
	Create(qrcode *QRCode) error
	GetByToken(token string) (*QRCode, error)
	GetByUserID(userID string) ([]*QRCode, error)
	Update(token string, url string) error
	Delete(token string) error
}
