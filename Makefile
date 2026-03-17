.PHONY: build install clean test lint fmt help

BINARY := rgm
BUILD_DIR := bin
GO_FILES := $(shell find . -name '*.go' -type f)

## Build
build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/rgm

install: build ## Install to GOPATH/bin
	go install ./cmd/rgm

## Development
fmt: ## Format code
	gofmt -s -w .

lint: ## Run linter
	golangci-lint run ./...

test: ## Run tests
	go test -v ./...

## Setup
deps: ## Download dependencies
	go mod tidy

init-config: build ## Initialize default config
	./$(BUILD_DIR)/$(BINARY) config init

## Cleanup
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
