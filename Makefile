# Makefile for GoMegle project

# Variables
BINARY_NAME=gomegle
BINARY_PATH=bin/$(BINARY_NAME)

# Default target
.DEFAULT_GOAL := build

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

# Build the application
build: bin
	go build -o $(BINARY_PATH) .

# Format the code
fmt:
	go fmt ./...

# Lint the code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Run the application (build first, then execute)
run: build
	./$(BINARY_PATH)

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

# Build for multiple platforms
build-all: bin
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_PATH)-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_PATH)-windows-amd64.exe .

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build the application to bin/gomegle"
	@echo "  fmt        - Format the Go source code"
	@echo "  lint       - Lint the Go source code (requires golangci-lint)"
	@echo "  run        - Build and run the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  test       - Run tests"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  help       - Show this help message"

# Declare phony targets
.PHONY: build fmt lint run clean deps test build-all help
