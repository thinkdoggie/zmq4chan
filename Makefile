# Makefile for zmq4chan

.PHONY: test test-verbose test-race bench clean examples lint format help

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run tests
	go test -v ./...

test-verbose: ## Run tests with verbose output
	go test -v -count=1 ./...

test-race: ## Run tests with race detection
	go test -race -v ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

examples: ## Run all examples
	@echo "Running REQ/REP example..."
	@cd examples/reqrep && go run main.go
	@echo ""
	@echo "Running PUB/SUB example..."
	@cd examples/pubsub && go run main.go

example-reqrep: ## Run REQ/REP example only
	@cd examples/reqrep && go run main.go

example-pubsub: ## Run PUB/SUB example only
	@cd examples/pubsub && go run main.go

lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format: ## Format code
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

clean: ## Clean up
	go clean -testcache
	go mod tidy

check: format lint test ## Run format, lint, and test

# Development helpers
dev-deps: ## Install development dependencies
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

mod-update: ## Update go modules
	go get -u ./...
	go mod tidy 