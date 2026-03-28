package usecase

import (
	"context"
	"fmt"

	"github.com/abinter/qr-code-generator/internal/domain"
	"github.com/abinter/qr-code-generator/pkg/token"
)

const maxURLLength = 2048
const maxRetries = 5

type QRCodeUseCase struct {
	repo    domain.QRCodeRepository
	baseURL string
}

func NewQRCodeUseCase(repo domain.QRCodeRepository, baseURL string) *QRCodeUseCase {
	return &QRCodeUseCase{repo: repo, baseURL: baseURL}
}

func (uc *QRCodeUseCase) CreateQRCode(ctx context.Context, userID, url string) (*domain.QRCode, error) {
	if err := validateURL(url); err != nil {
		return nil, err
	}

	for range maxRetries {
		tok := token.Generate(url)
		qr := &domain.QRCode{
			UserID:  userID,
			QRToken: tok,
			URL:     url,
		}

		err := uc.repo.Create(ctx, qr)
		if err == nil {
			return qr, nil
		}
		if err != domain.ErrTokenConflict {
			return nil, fmt.Errorf("create qr code: %w", err)
		}
	}

	return nil, fmt.Errorf("create qr code: exceeded max retries on token collision")
}

func (uc *QRCodeUseCase) GetQRCode(ctx context.Context, qrToken string) (*domain.QRCode, error) {
	qr, err := uc.repo.GetByToken(ctx, qrToken)
	if err != nil {
		return nil, fmt.Errorf("get qr code: %w", err)
	}
	return qr, nil
}

func (uc *QRCodeUseCase) EditQRCode(ctx context.Context, qrToken, newURL string) error {
	if err := validateURL(newURL); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, qrToken, newURL); err != nil {
		return fmt.Errorf("edit qr code: %w", err)
	}
	return nil
}

func (uc *QRCodeUseCase) DeleteQRCode(ctx context.Context, qrToken string) error {
	if err := uc.repo.Delete(ctx, qrToken); err != nil {
		return fmt.Errorf("delete qr code: %w", err)
	}
	return nil
}

func (uc *QRCodeUseCase) ListQRCodes(ctx context.Context, userID string) ([]*domain.QRCode, error) {
	qrs, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list qr codes: %w", err)
	}
	return qrs, nil
}

func validateURL(url string) error {
	if url == "" {
		return domain.ErrInvalidURL
	}
	if len(url) > maxURLLength {
		return domain.ErrURLTooLong
	}
	return nil
}
