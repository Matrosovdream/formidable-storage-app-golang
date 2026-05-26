.PHONY: help dev dev-down dev-logs prod prod-down prod-deploy build test tidy migrate-up migrate-down seed

DEV  := docker compose -f compose.dev.yaml
PROD := docker compose -f compose.prod.yaml

DB_URL ?= postgres://laravel:secret@localhost:5432/app?sslmode=disable

help:
	@echo "Targets:"
	@echo "  dev          - bring up dev stack (postgres, redis, web on :8080, worker)"
	@echo "  dev-down     - stop dev stack"
	@echo "  dev-logs     - tail web + worker logs"
	@echo "  prod         - bring up prod stack (requires .env)"
	@echo "  prod-down    - stop prod stack"
	@echo "  prod-deploy  - memory-safe redeploy (sequential builds, for small VMs)"
	@echo "  seed         - seed whichever stack is running (dev or prod)"
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
	$(PROD) --env-file .env up -d --build

prod-down:
	$(PROD) down

# Memory-safe redeploy for small droplets: builds services one at a time so the
# Go compiler doesn't try to build web + worker in parallel and OOM the host.
prod-deploy:
	$(PROD) build web
	$(PROD) build worker
	$(PROD) --env-file .env up -d

# Seeds whichever stack is currently running (prod takes precedence if both).
seed:
	@if [ -n "$$($(PROD) ps -q postgres 2>/dev/null)" ]; then \
		echo ">>> Seeding prod stack"; \
		$(PROD) --profile seed run --rm --build seed; \
	elif [ -n "$$($(DEV) ps -q postgres 2>/dev/null)" ]; then \
		echo ">>> Seeding dev stack"; \
		$(DEV) --profile seed run --rm seed; \
	else \
		echo "No stack running. Start one with 'make dev' or 'make prod' first."; \
		exit 1; \
	fi

build:
	go build ./...

test:
	go test ./...

tidy:
	go mod tidy

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1
