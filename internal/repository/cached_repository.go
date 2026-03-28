package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/abinter/qr-code-generator/internal/domain"
)

const cacheKeyPrefix = "qr:token:"

// CachedRepository wraps a QRCodeRepository and adds a Redis read-through cache.
// Cache strategy:
//   - GetByToken: cache-aside (miss → DB → populate cache)
//   - Update / Delete: write-through (DB first, then invalidate cache)
//   - Create / GetByUserID: pass-through (no caching)
type CachedRepository struct {
	repo   domain.QRCodeRepository
	client *redis.Client
	ttl    time.Duration
}

// Ensure CachedRepository implements domain.QRCodeRepository at compile time.
var _ domain.QRCodeRepository = (*CachedRepository)(nil)

func NewCachedRepository(repo domain.QRCodeRepository, client *redis.Client, ttl time.Duration) *CachedRepository {
	return &CachedRepository{repo: repo, client: client, ttl: ttl}
}

func (c *CachedRepository) Create(ctx context.Context, qrcode *domain.QRCode) error {
	return c.repo.Create(ctx, qrcode)
}

func (c *CachedRepository) GetByToken(ctx context.Context, token string) (*domain.QRCode, error) {
	key := cacheKey(token)

	// Cache hit
	data, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		var qr domain.QRCode
		if jsonErr := json.Unmarshal(data, &qr); jsonErr == nil {
			return &qr, nil
		}
	}

	// Cache miss — fetch from DB
	qr, err := c.repo.GetByToken(ctx, token)
	if err != nil {
		// Do not cache not-found or errors
		return nil, err
	}

	// Populate cache; ignore cache write errors (best-effort)
	if data, jsonErr := json.Marshal(qr); jsonErr == nil {
		c.client.Set(ctx, key, data, c.ttl)
	}

	return qr, nil
}

func (c *CachedRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.QRCode, error) {
	return c.repo.GetByUserID(ctx, userID)
}

func (c *CachedRepository) Update(ctx context.Context, token string, url string) error {
	if err := c.repo.Update(ctx, token, url); err != nil {
		return err
	}
	c.invalidate(ctx, token)
	return nil
}

func (c *CachedRepository) Delete(ctx context.Context, token string) error {
	if err := c.repo.Delete(ctx, token); err != nil {
		return err
	}
	c.invalidate(ctx, token)
	return nil
}

func (c *CachedRepository) invalidate(ctx context.Context, token string) {
	if err := c.client.Del(ctx, cacheKey(token)).Err(); err != nil && !errors.Is(err, redis.Nil) {
		// Log but don't fail the operation — cache is best-effort
		fmt.Printf("cache invalidation warning for token %q: %v\n", token, err)
	}
}

func cacheKey(token string) string {
	return cacheKeyPrefix + token
}
