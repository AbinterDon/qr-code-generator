package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/abinter/qr-code-generator/internal/domain"
)

// Ensure PostgresRepository implements domain.QRCodeRepository at compile time.
var _ domain.QRCodeRepository = (*PostgresRepository)(nil)

const pgUniqueViolation = "23505"

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, qrcode *domain.QRCode) error {
	query := `
		INSERT INTO qr_codes (user_id, qr_token, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	err := r.pool.QueryRow(ctx, query,
		qrcode.UserID, qrcode.QRToken, qrcode.URL, now,
	).Scan(&qrcode.ID, &qrcode.CreatedAt, &qrcode.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrTokenConflict
		}
		return fmt.Errorf("create qr code: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByToken(ctx context.Context, token string) (*domain.QRCode, error) {
	query := `
		SELECT id, user_id, qr_token, url, created_at, updated_at
		FROM qr_codes
		WHERE qr_token = $1`

	var qr domain.QRCode
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&qr.ID, &qr.UserID, &qr.QRToken, &qr.URL, &qr.CreatedAt, &qr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get qr code by token: %w", err)
	}
	return &qr, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.QRCode, error) {
	query := `
		SELECT id, user_id, qr_token, url, created_at, updated_at
		FROM qr_codes
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get qr codes by user: %w", err)
	}
	defer rows.Close()

	var results []*domain.QRCode
	for rows.Next() {
		var qr domain.QRCode
		if err := rows.Scan(&qr.ID, &qr.UserID, &qr.QRToken, &qr.URL, &qr.CreatedAt, &qr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan qr code row: %w", err)
		}
		results = append(results, &qr)
	}
	return results, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, token string, url string) error {
	query := `
		UPDATE qr_codes
		SET url = $1, updated_at = $2
		WHERE qr_token = $3`

	tag, err := r.pool.Exec(ctx, query, url, time.Now(), token)
	if err != nil {
		return fmt.Errorf("update qr code: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM qr_codes WHERE qr_token = $1`

	tag, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete qr code: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
