# Makefile for zmq4chan

.PHONY: test test-verbose test-race bench clean examples lint format help
.PHONY: test-v1 test-v2 test-all lint-v1 lint-v2 lint-all examples-v1 examples-v2 examples-all
.PHONY: bench-v1 bench-v2 bench-all format-v1 format-v2 format-all clean-v1 clean-v2 clean-all

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Combined targets (run for both v1 and v2)
test: test-all ## Run tests for both v1 and v2

test-all: test-v1 test-v2 ## Run tests for both versions

test-verbose: test-v1-verbose test-v2-verbose ## Run verbose tests for both versions

test-race: test-v1-race test-v2-race ## Run race tests for both versions

bench: bench-all ## Run benchmarks for both versions

bench-all: bench-v1 bench-v2 ## Run benchmarks for both versions

examples: examples-all ## Run examples for both versions

examples-all: examples-v1 examples-v2 ## Run examples for both versions

lint: lint-all ## Run linter for both versions

lint-all: lint-v1 lint-v2 ## Run linter for both versions

format: format-all ## Format code for both versions

format-all: format-v1 format-v2 ## Format code for both versions

clean: clean-all ## Clean up for both versions

clean-all: clean-v1 clean-v2 ## Clean up for both versions

# V1 specific targets
test-v1: ## Run v1 tests
	go test -v ./...

test-v1-verbose: ## Run v1 tests with verbose output
	go test -v -count=1 ./...

test-v1-race: ## Run v1 tests with race detection
	go test -race -v ./...

bench-v1: ## Run v1 benchmarks
	go test -bench=. -benchmem ./...

examples-v1: ## Run all v1 examples
	@echo "Running v1 REQ/REP example..."
	@cd examples/reqrep && go run main.go
	@echo ""
	@echo "Running v1 PUB/SUB example..."
	@cd examples/pubsub && go run main.go

example-v1-reqrep: ## Run v1 REQ/REP example only
	@cd examples/reqrep && go run main.go

example-v1-pubsub: ## Run v1 PUB/SUB example only
	@cd examples/pubsub && go run main.go

lint-v1: ## Run v1 linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format-v1: ## Format v1 code
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

clean-v1: ## Clean up v1
	go clean -testcache
	go mod tidy

# V2 specific targets
test-v2: ## Run v2 tests
	cd v2 && go test -v ./...

test-v2-verbose: ## Run v2 tests with verbose output
	cd v2 && go test -v -count=1 ./...

test-v2-race: ## Run v2 tests with race detection
	cd v2 && go test -race -v ./...

bench-v2: ## Run v2 benchmarks
	cd v2 && go test -bench=. -benchmem ./...

examples-v2: ## Run all v2 examples
	@echo "Running v2 REQ/REP example..."
	@cd v2/examples/reqrep && go run main.go
	@echo ""
	@echo "Running v2 PUB/SUB example..."
	@cd v2/examples/pubsub && go run main.go

example-v2-reqrep: ## Run v2 REQ/REP example only
	@cd v2/examples/reqrep && go run main.go

example-v2-pubsub: ## Run v2 PUB/SUB example only
	@cd v2/examples/pubsub && go run main.go

lint-v2: ## Run v2 linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd v2 && golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format-v2: ## Format v2 code
	cd v2 && go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		cd v2 && goimports -w .; \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

clean-v2: ## Clean up v2
	cd v2 && go clean -testcache
	cd v2 && go mod tidy

# Version comparison targets
compare-bench: ## Compare benchmarks between v1 and v2
	@echo "Running v1 benchmarks..."
	@go test -bench=. -benchmem ./... > bench-v1.txt 2>&1 || true
	@echo ""
	@echo "Running v2 benchmarks..."
	@cd v2 && go test -bench=. -benchmem ./... > ../bench-v2.txt 2>&1 || true
	@echo ""
	@echo "Benchmark results saved to bench-v1.txt and bench-v2.txt"
	@echo "Use 'benchcmp bench-v1.txt bench-v2.txt' to compare (install with: go install golang.org/x/tools/cmd/benchcmp@latest)"

# Development helpers
dev-deps: ## Install development dependencies
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/tools/cmd/benchcmp@latest

mod-update: ## Update go modules for both versions
	go get -u ./...
	go mod tidy
	cd v2 && go get -u ./...
	cd v2 && go mod tidy

mod-update-v1: ## Update v1 go modules
	go get -u ./...
	go mod tidy 

mod-update-v2: ## Update v2 go modules
	cd v2 && go get -u ./...
	cd v2 && go mod tidy

check: check-all ## Run format, lint, and test for both versions

check-all: format-all lint-all test-all ## Run format, lint, and test for both versions

check-v1: format-v1 lint-v1 test-v1 ## Run format, lint, and test for v1

check-v2: format-v2 lint-v2 test-v2 ## Run format, lint, and test for v2 