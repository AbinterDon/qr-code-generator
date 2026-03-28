package domain

import "context"

// MockQRCodeRepository is a test double for QRCodeRepository.
type MockQRCodeRepository struct {
	CreateFunc      func(ctx context.Context, qrcode *QRCode) error
	GetByTokenFunc  func(ctx context.Context, token string) (*QRCode, error)
	GetByUserIDFunc func(ctx context.Context, userID string) ([]*QRCode, error)
	UpdateFunc      func(ctx context.Context, token string, url string) error
	DeleteFunc      func(ctx context.Context, token string) error
}

func (m *MockQRCodeRepository) Create(ctx context.Context, qrcode *QRCode) error {
	return m.CreateFunc(ctx, qrcode)
}

func (m *MockQRCodeRepository) GetByToken(ctx context.Context, token string) (*QRCode, error) {
	return m.GetByTokenFunc(ctx, token)
}

func (m *MockQRCodeRepository) GetByUserID(ctx context.Context, userID string) ([]*QRCode, error) {
	return m.GetByUserIDFunc(ctx, userID)
}

func (m *MockQRCodeRepository) Update(ctx context.Context, token string, url string) error {
	return m.UpdateFunc(ctx, token, url)
}

func (m *MockQRCodeRepository) Delete(ctx context.Context, token string) error {
	return m.DeleteFunc(ctx, token)
}
