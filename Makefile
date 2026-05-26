.PHONY: help dev dev-down dev-logs prod prod-down build test tidy migrate-up migrate-down seed seed-rebuild

DEV  := docker compose -f compose.dev.yaml
PROD := docker compose -f compose.prod.yaml

DB_URL ?= postgres://laravel:secret@localhost:5432/app?sslmode=disable

help:
	@echo "Targets:"
	@echo "  dev          - bring up dev stack (postgres, redis, web on :8080, worker)"
	@echo "  dev-down     - stop dev stack"
	@echo "  dev-logs     - tail web + worker logs"
	@echo "  prod         - deploy / redeploy prod stack (sequential builds, safe on small VMs)"
	@echo "  prod-down    - stop prod stack"
	@echo "  seed         - seed whichever stack is running (dev or prod)"
	@echo "  seed-rebuild - rebuild the prod seed image (after editing cmd/seed/)"
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

# Deploys / redeploys the prod stack. Builds services sequentially so the Go
# compiler doesn't run two compiles in parallel and OOM small VMs (a 2 vCPU /
# 4 GB droplet can't handle parallel builds even with BuildKit cache mounts).
prod:
	$(PROD) build web
	$(PROD) build worker
	$(PROD) --env-file .env up -d

prod-down:
	$(PROD) down

# Seeds whichever stack is currently running (prod takes precedence if both).
# First invocation builds the seed image; subsequent ones reuse it (fast).
# Run `make seed-rebuild` after editing cmd/seed/ or its embedded JSON.
seed:
	@if [ -n "$$($(PROD) ps -q postgres 2>/dev/null)" ]; then \
		echo ">>> Seeding prod stack"; \
		$(PROD) --profile seed run --rm seed; \
	elif [ -n "$$($(DEV) ps -q postgres 2>/dev/null)" ]; then \
		echo ">>> Seeding dev stack"; \
		$(DEV) --profile seed run --rm seed; \
	else \
		echo "No stack running. Start one with 'make dev' or 'make prod' first."; \
		exit 1; \
	fi

# Force a rebuild of the seed image (use after editing cmd/seed/seeds/*.json
# or the seed binary source). Cheap thanks to BuildKit cache mounts.
seed-rebuild:
	@if [ -n "$$($(PROD) ps -q postgres 2>/dev/null)" ]; then \
		$(PROD) --profile seed build seed; \
	else \
		echo "Prod stack not running — dev seed uses 'go run' and never needs a rebuild."; \
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
