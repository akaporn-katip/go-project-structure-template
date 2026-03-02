
# .PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Build
build:
	go build -o bin/api cmd/api/main.go

# Run
run:
	go run cmd/api/main.go

# Test
test:
	go test -v -race ./...


test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	go clean

# Install tools
install-tools:
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest

deps:
	go mod download
	go mod tidy
