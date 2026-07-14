.PHONY: build run migrate generate fmt tidy

APP_NAME := go-api-boilerplate
MIGRATIONS_DIR := internal/db/migrations

## build: Compile the binary
build:
	go build -o bin/$(APP_NAME) ./cmd/api

## run: Run the server locally
run:
	go run ./cmd/api

## migrate: Apply all pending migrations using goose CLI
migrate:
	goose up

## generate: Run sqlc to regenerate Go code from SQL
generate:
	sqlc generate

## fmt: Format code
fmt:
	gofmt -s -w .
	goimports -w .

## tidy: Tidy module dependencies
tidy:
	go mod tidy

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'