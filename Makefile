.PHONY: help deps fmt test test-integration build run \
	up down rebuild logs ps \
	migrate seed-stats clean

# ---- Common settings ---------------------------------------------------------

GO ?= go
APP_MAIN ?= ./cmd/tracker-bot/main.go
MIGRATOR_MAIN ?= ./cmd/migrator/main.go
SEED_MAIN ?= ./cmd/seed-stats/main.go

GOCACHE_DIR ?= $(CURDIR)/.gocache
COMPOSE ?= docker compose

# For local DB tools/clients.
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5438
DB_NAME ?= tracker
DB_USER ?= tracker
DB_PASS ?= tracker

help: ## Show available commands
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-14s %s\n", $$1, $$2}'

deps: ## Download and tidy Go modules
	$(GO) mod download
	$(GO) mod tidy

fmt: ## Format Go files
	$(GO) fmt ./...

test: ## Run unit tests (fast, no DB required)
	GOCACHE=$(GOCACHE_DIR) $(GO) test ./...

test-integration: ## Run integration tests against the local docker-compose DB (needs `make up && make migrate` first)
	GOCACHE=$(GOCACHE_DIR) TEST_DB_HOST=$(DB_HOST) TEST_DB_PORT=$(DB_PORT) TEST_DB_USER=$(DB_USER) TEST_DB_PASS=$(DB_PASS) TEST_DB_NAME=$(DB_NAME) \
		$(GO) test -tags=integration ./...

build: ## Build tracker bot binary
	mkdir -p bin
	$(GO) build -o bin/tracker-bot $(APP_MAIN)

run: ## Run tracker bot locally
	$(GO) run $(APP_MAIN)

up: ## Start services with Docker Compose
	$(COMPOSE) up -d

down: ## Stop services
	$(COMPOSE) down

rebuild: ## Rebuild and recreate services
	$(COMPOSE) up -d --build --force-recreate

logs: ## Follow docker compose logs
	$(COMPOSE) logs -f

ps: ## Show docker compose status
	$(COMPOSE) ps

migrate: ## Run DB migrations via Go migrator
	$(GO) run $(MIGRATOR_MAIN)

seed-stats: ## Seed demo sessions (requires TG user id)
	@if [ -z "$(TG_USER_ID)" ]; then \
		echo "Usage: make seed-stats TG_USER_ID=<telegram_user_id>"; \
		exit 1; \
	fi
	$(GO) run $(SEED_MAIN) -tg-user-id $(TG_USER_ID)

clean: ## Remove local build/cache artifacts
	rm -rf bin .gocache
