.PHONY: help dev dev-down dev-logs prod prod-down build test tidy migrate-up migrate-down

DEV  := docker compose -f compose.dev.yaml
PROD := docker compose -f compose.prod.yaml

DB_URL ?= postgres://laravel:secret@localhost:5432/app?sslmode=disable

help:
	@echo "Targets:"
	@echo "  dev          - bring up dev stack (postgres, redis, web on :8080, worker)"
	@echo "  dev-down     - stop dev stack"
	@echo "  dev-logs     - tail web + worker logs"
	@echo "  prod         - bring up prod stack (requires .env.prod)"
	@echo "  prod-down    - stop prod stack"
	@echo "  build        - go build ./..."
	@echo "  test         - go test ./..."
	@echo "  tidy         - go mod tidy"
	@echo "  migrate-up   - apply pending migrations (DB_URL=...)"
	@echo "  migrate-down - revert last migration"

dev:
	$(DEV) up --build

dev-down:
	$(DEV) down

dev-logs:
	$(DEV) logs -f web worker

prod:
	$(PROD) --env-file .env.prod up -d --build

prod-down:
	$(PROD) down

build:
	go build ./...

test:
	go test ./...

tidy:
	go mod tidy

seed:
	DB_HOST=localhost DB_PORT=5432 DB_DATABASE=app DB_USERNAME=laravel DB_PASSWORD=secret DB_SSLMODE=disable REDIS_HOST=localhost go run ./cmd/seed

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1
