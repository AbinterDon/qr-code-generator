run:
	go run ./cmd/server

test:
	go test -race ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

tidy:
	go mod tidy

build:
	go build -o bin/server ./cmd/server
