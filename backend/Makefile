# Simple Makefile for backwards compatibility
# Use Task for full feature set: task --list

.PHONY: test test-coverage lint

test:
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

run:
	@go run cmd/server/main.go

.DEFAULT_GOAL := test
