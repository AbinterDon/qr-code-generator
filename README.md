# QR Code Generator

A production-ready QR Code generation and management service built in Go.

Supports creating QR codes from URLs, retrieving QR code images, scanning to redirect, and managing codes per user. Designed for scale (1B QR codes, 100M users).

## Features

- **Create** a QR code from any URL → returns a unique `qr_token`
- **Scan** a QR code → 302 redirect to the original URL
- **Retrieve** the QR code image (customizable size, color, border)
- **Edit / Delete** QR codes
- **Redis cache** for read-heavy workloads (optional)
- **PostgreSQL** for persistent storage

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| HTTP Router | [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL 16 |
| Cache | Redis 7 (optional) |
| QR Image | [go-qrcode](https://github.com/skip2/go-qrcode) |

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

## Quick Start

**1. Clone the repo**

```bash
git clone https://github.com/AbinterDon/qr-code-generator.git
cd qr-code-generator
```

**2. Start PostgreSQL and Redis**

```bash
docker compose up -d
```

**3. Run the server**

```bash
DATABASE_URL="postgres://qr:qr@localhost:5432/qrcode?sslmode=disable" \
REDIS_URL="redis://localhost:6379" \
go run ./cmd/server
```

The server starts on `http://localhost:8080`. Database migrations run automatically on startup.

## API

### Create a QR code

```bash
curl -X POST http://localhost:8080/v1/qr_code \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

```json
{ "qr_token": "aB3kR7zQ" }
```

### Get QR code image

```
GET /v1/qr_code_image/:qr_token?dimension=256&color=000000&border=4
```

```bash
curl "http://localhost:8080/v1/qr_code_image/aB3kR7zQ?dimension=300" \
  --output qr.png
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `dimension` | int | `256` | Image size in pixels |
| `color` | hex string | `000000` | Foreground color (e.g. `FF0000` for red) |
| `border` | int | `4` | White border width in pixels |

### Scan (redirect)

```
GET /:qr_token   →   302 redirect to original URL
```

```bash
curl -L http://localhost:8080/aB3kR7zQ
```

### Get original URL

```bash
curl http://localhost:8080/v1/qr_code/aB3kR7zQ
```

```json
{ "url": "https://example.com" }
```

### Edit a QR code

```bash
curl -X PUT http://localhost:8080/v1/qr_code/aB3kR7zQ \
  -H "Content-Type: application/json" \
  -d '{"url": "https://new-url.com"}'
```

### Delete a QR code

```bash
curl -X DELETE http://localhost:8080/v1/qr_code/aB3kR7zQ
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `REDIS_URL` | No | — | Redis connection string. Cache is disabled if unset |
| `SERVER_ADDR` | No | `:8080` | Address the server listens on |
| `BASE_URL` | No | `http://localhost:8080` | Public base URL (embedded in QR codes) |

## Running Tests

```bash
# Unit tests (no external dependencies)
go test ./...

# Unit tests with race detector
go test -race ./...

# Integration tests (requires Docker)
docker compose up -d
go test -tags=integration ./internal/repository/...
```

## Project Structure

```
qr-code-generator/
├── cmd/server/         # Entry point
├── db/                 # Connection pool + embedded migrations
│   └── migrations/
├── internal/
│   ├── config/         # Environment config
│   ├── domain/         # QRCode entity, repository interface, errors
│   ├── handler/        # HTTP handlers and chi router
│   ├── repository/     # PostgreSQL + Redis cache implementations
│   └── usecase/        # Business logic (create, get, edit, delete)
├── pkg/
│   ├── qrimage/        # PNG image generator
│   └── token/          # SHA-256 + Base62 token generator
└── docker-compose.yml  # PostgreSQL + Redis
```

## Architecture

This project follows **Clean Architecture**. Dependencies point inward:

```
handler → usecase → domain ← repository
```

- `domain` defines the `QRCode` entity and `QRCodeRepository` interface
- `usecase` contains all business logic; depends only on the interface
- `repository` has two implementations that both satisfy the interface:
  - `PostgresRepository` — persistent storage
  - `CachedRepository` — Redis decorator wrapping any repository
- `handler` translates HTTP ↔ use case calls

This means you can swap any layer independently (e.g. replace Redis with Memcached, or swap PostgreSQL for another DB) without touching business logic.

## Token Generation

Tokens are generated as follows:

1. SHA-256 hash of `url + random nonce`
2. Encode with Base62 (`[0-9A-Za-z]`)
3. Take the first 8 characters as `qr_token`
4. PostgreSQL `UNIQUE` constraint guarantees uniqueness; collisions trigger a retry

## Caching Strategy

The `CachedRepository` uses **cache-aside** for reads and **write-through invalidation** for writes:

- `GetByToken` — check Redis first; on miss, fetch from DB and populate cache
- `Update` / `Delete` — write to DB first, then delete the cache key
- Not-found results are **not** cached
- Default TTL: 24 hours
