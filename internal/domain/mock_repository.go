package domain

// MockQRCodeRepository is a test double for QRCodeRepository.
type MockQRCodeRepository struct {
	CreateFunc      func(qrcode *QRCode) error
	GetByTokenFunc  func(token string) (*QRCode, error)
	GetByUserIDFunc func(userID string) ([]*QRCode, error)
	UpdateFunc      func(token string, url string) error
	DeleteFunc      func(token string) error
}

func (m *MockQRCodeRepository) Create(qrcode *QRCode) error {
	return m.CreateFunc(qrcode)
}

func (m *MockQRCodeRepository) GetByToken(token string) (*QRCode, error) {
	return m.GetByTokenFunc(token)
}

func (m *MockQRCodeRepository) GetByUserID(userID string) ([]*QRCode, error) {
	return m.GetByUserIDFunc(userID)
}

func (m *MockQRCodeRepository) Update(token string, url string) error {
	return m.UpdateFunc(token, url)
}

func (m *MockQRCodeRepository) Delete(token string) error {
	return m.DeleteFunc(token)
}
