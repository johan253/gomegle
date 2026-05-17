# Makefile for GoMegle project

# Variables
BINARY_NAME=gomegle
BINARY_PATH=bin/$(BINARY_NAME)
HELM_TAG := $(shell helm show chart helm/ | grep '^version:' | awk '{print $$2}')
GIT_SHA := $(shell git rev-parse --short HEAD)
IMAGE_TAG := $(HELM_TAG)-$(GIT_SHA)

# Default target
.DEFAULT_GOAL := build

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

# Build the protobuf files
proto:
	protoc --go_out=. --go_opt=paths=source_relative models.proto

# Build the application
build: deps proto bin
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

# Development mode with auto-restart on file changes
dev:
	./scripts/dev.sh

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

# Run test to test server runs
test-server:
	./scripts/test_server_run.sh

# Build for multiple platforms
build-all: bin helm-build
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_PATH)-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_PATH)-windows-amd64.exe .

# Print the current image tag
tag:
	@echo $(IMAGE_TAG)

helm-lint:
	helm lint helm/

helm-build: helm-lint
	helm package helm/ --version $(IMAGE_TAG)

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build the application to bin/gomegle"
	@echo "  fmt        - Format the Go source code"
	@echo "  lint       - Lint the Go source code (requires golangci-lint)"
	@echo "  run        - Build and run the application"
	@echo "  dev        - Run in development mode with auto-restart on file changes"
	@echo "  clean      - Remove build artifacts"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  test       - Run tests"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  tag        - Print the current image tag (<chart_tag>-<git_sha>)"
	@echo "  help       - Show this help message"

# Declare phony targets
.PHONY: build fmt lint run dev clean deps test build-all help proto image-tag helm-lint helm-build
