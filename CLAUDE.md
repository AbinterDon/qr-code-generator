# QR Code Generator

A QR Code generation and management service built in Go, designed for scale (1B QR codes, 100M users).

## Project Overview

Based on the system design in `1_Design a QR code Generator.pdf`:
- Create QR codes from URLs (returns a unique `qr_token`)
- Retrieve QR code images with customizable dimensions/color/border
- Scan to redirect (302) to the original URL
- Manage (edit, delete) QR codes per user

## Tech Stack

- **Language**: Go 1.22+
- **Architecture**: Clean Architecture (domain / usecase / adapter / infrastructure)
- **Database**: Relational DB with UNIQUE index on `qr_token`
- **Cache**: Redis (read-heavy, write:read ≈ 1:100)
- **CDN**: Static QR code images served via CDN

## Skills

Apply the relevant skill when working in each context:

| Context | Skill |
|---------|-------|
| Writing Go code | `.agent/skills/golang-patterns/SKILL.md` |
| Writing tests | `.agent/skills/golang-testing/SKILL.md` |
| Designing architecture / layers | `.agent/skills/architecture-patterns/SKILL.md` |

## Project Structure

```
qr-code-generator/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/         # QrCode entity, Repository interface
│   ├── usecase/        # Create, get, edit, delete QR codes
│   ├── handler/        # HTTP handlers
│   ├── repository/     # DB implementation
│   └── config/
├── pkg/                # Public utilities (e.g. token generator)
├── api/                # OpenAPI / route definitions
├── testdata/
├── go.mod
└── Makefile
```

## API Design

```
POST   /v1/qr_code                      Create a QR code
GET    /v1/qr_code_image/:qr_token      Get QR code image (?dimension=&color=&border=)
GET    /v1/qr_code/:qr_token            Get original URL
PUT    /v1/qr_code/:qr_token            Edit QR code
DELETE /v1/qr_code/:qr_token            Delete QR code
GET    /:qr_token                        Scan → 302 redirect to original URL
```

## Token Generation

- SHA-256 hash of URL + nonce, encoded with Base62
- Take first N characters as `qr_token`
- DB enforces UNIQUE constraint; regenerate on collision

## Coding Standards

- Follow TDD: write tests before implementation
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- No global variables; use dependency injection
- `context.Context` is always the first function parameter
- Always use 302 (Temporary Redirect) for QR code scans

## Common Commands

```bash
go run ./cmd/server        # Start the server
go test ./...              # Run all tests
go test -race ./...        # Run with race detector
go test -cover ./...       # Run with coverage
golangci-lint run          # Lint
go mod tidy                # Tidy dependencies
```
