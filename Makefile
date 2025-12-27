.PHONY: build build-all test clean run-stdio run-http docker-build docker-run

VERSION ?= 1.0.0
BUILD_FLAGS = -ldflags "-X github.com/jeff-french/ynab-mcp-server/cmd.Version=$(VERSION)"

# Build for current platform
build:
	@echo "Building ynab-mcp-server..."
	go build $(BUILD_FLAGS) -o ynab-mcp-server

# Cross-compile for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/ynab-mcp-server-linux-amd64
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/ynab-mcp-server-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o dist/ynab-mcp-server-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/ynab-mcp-server-windows-amd64.exe
	@echo "Built all platforms successfully!"
	@ls -lh dist/

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist ynab-mcp-server coverage.txt coverage.html

# Run in stdio mode (for development/testing)
run-stdio:
	@echo "Starting in stdio mode..."
	go run . serve --transport=stdio

# Run in HTTP mode (for development/testing)
run-http:
	@echo "Starting in HTTP mode on port 8080..."
	go run . serve --transport=http --port=8080

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t ynab-mcp-server:$(VERSION) .
	docker tag ynab-mcp-server:$(VERSION) ynab-mcp-server:latest

# Run with Docker Compose
docker-run:
	docker-compose up

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "YNAB MCP Server - Makefile commands:"
	@echo ""
	@echo "  make build          - Build for current platform"
	@echo "  make build-all      - Cross-compile for all platforms"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run-stdio      - Run in stdio mode (development)"
	@echo "  make run-http       - Run in HTTP mode (development)"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run with Docker Compose"
	@echo "  make deps           - Install and tidy dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code"
	@echo "  make help           - Show this help message"
