APP_NAME := repurpose-be
BIN_DIR  := bin
BINARY   := $(BIN_DIR)/$(APP_NAME)

.PHONY: help run worker build test vet fmt tidy lint migrate-up migrate-down docker-build docker-up docker-down

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

run: ## Run HTTP API server
	go run . server

worker: ## Run background worker
	go run . worker

build: ## Build binary into bin/
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -o $(BINARY) .

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format source files
	gofmt -w .

tidy: ## Sync go.mod/go.sum
	go mod tidy

lint: fmt vet ## Format + vet (placeholder for golangci-lint)

migrate-up: ## Apply all database migrations (PostgreSQL)
	DB_DRIVER=postgres go run . migrate --direction=up

migrate-down: ## Roll back the last migration (PostgreSQL)
	DB_DRIVER=postgres go run . migrate --direction=down --step=1

docker-build: ## Build Docker images
	docker compose build

docker-up: ## Start full stack (postgres, redis, api, worker)
	docker compose up --build

docker-down: ## Stop full stack and remove volumes
	docker compose down -v
