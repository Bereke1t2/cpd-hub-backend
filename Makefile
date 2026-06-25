DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/cpdhub?sslmode=disable
MIGRATIONS_DIR ?= migrations

.PHONY: help run build test vet fmt tidy docker-build migrate-up migrate-down seed

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

run: ## Run the server locally
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/server

build: ## Build the server binary into ./bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

test: ## Run all tests
	go test ./... -count=1

vet: ## go vet
	go vet ./...

fmt: ## Format all code
	gofmt -w .

tidy: ## Tidy modules
	go mod tidy

docker-build: ## Build the production image
	docker build -t cpd-hub-backend:latest .

migrate-up: ## Apply all up migrations
	go run ./cmd/migrate -dir $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down: ## Roll back the last migration
	go run ./cmd/migrate -dir $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

seed: ## Load dev seed data
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed
