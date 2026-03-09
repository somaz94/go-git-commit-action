.PHONY: build test test-unit test-all cover cover-html bench lint fmt clean help

BINARY := go-git-commit-action
GOFLAGS := -v

## Build

build: ## Build the binary
	go build $(GOFLAGS) -o $(BINARY) ./cmd/main.go

## Test

test: test-unit ## Run unit tests (alias)

test-unit: ## Run unit tests with coverage
	go test ./internal/... ./cmd/... -v -cover

test-all: test-unit ## Run all tests

## Coverage

cover: ## Generate coverage report
	go test ./internal/... ./cmd/... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html: cover ## Open coverage report in browser
	go tool cover -html=coverage.out

## Benchmark

bench: ## Run benchmarks
	go test -bench=. -benchmem ./internal/...

## Quality

lint: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofmt -s -w .

## Cleanup

clean: ## Remove build artifacts and coverage files
	rm -f $(BINARY) coverage.out

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
