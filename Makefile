.PHONY: build run test clean install deps

# Build variables
BINARY_NAME=onefirewall-classifier
BUILD_DIR=build
GO=go
GOFLAGS=-v

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with custom port
run-port: build
	@echo "Running $(BINARY_NAME) on port 9090..."
	./$(BUILD_DIR)/$(BINARY_NAME) -port 9090

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Run the application in development mode with sample data
dev: build
	@echo "Running in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) -log-level debug

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/server/main.go

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):latest

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  run-port      - Build and run on custom port (9090)"
	@echo "  deps          - Install dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  dev           - Run in development mode"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
