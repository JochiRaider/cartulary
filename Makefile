.PHONY: bootstrap db-up db-reset dev generate generate-drift migration-drift test lint check build

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
PNPM ?= $(shell if command -v pnpm >/dev/null 2>&1; then command -v pnpm; elif [ -x "$$HOME/.local/share/pnpm/pnpm" ]; then printf "$$HOME/.local/share/pnpm/pnpm"; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)

SQLC_TOOL := github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
GOOSE_TOOL := github.com/pressly/goose/v3/cmd/goose@v3.27.0
TESTCONTAINERS_GO_VERSION := v0.42.0

bootstrap:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) install $(SQLC_TOOL)
	$(GO_RUN_ENV) $(GO) install $(GOOSE_TOOL)
	$(PNPM) install

db-up:
	docker compose -f docker-compose.dev.yml up -d postgres minio

db-reset:
	docker compose -f docker-compose.dev.yml up -d postgres
	docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS cartulary;"
	docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "CREATE DATABASE cartulary;"
	CARTULARY_CONFIG_FILE=$(CONFIG_FILE) $(GO_RUN_ENV) $(GO) run ./cmd/migrate up

dev:
	CARTULARY_CONFIG_FILE=$(CONFIG_FILE) $(GO_RUN_ENV) $(GO) run ./cmd/server & \
	$(PNPM) --dir apps/web dev & \
	wait

generate:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) run ./tools/contractgen
	@printf '%s\n' 'generate: sqlc skipped; no sqlc config or authored query generation inputs are present in this slice.'

# Codegen drift is distinct from migration drift.
generate-drift:
	./scripts/check-generate-drift.sh

# Migration drift covers schema-affecting changes not represented in /db/migrations
# or migrations that fail to apply cleanly in CI.
migration-drift:
	GO=$(GO) CONFIG_FILE=$(CONFIG_FILE) GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR) ./scripts/check-migrations.sh

test:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) test ./...
	$(PNPM) --dir apps/web test --run --passWithNoTests

lint:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) vet ./...
	$(PNPM) --dir apps/web exec biome check src
	$(PNPM) --dir apps/web typecheck

check:
	$(PNPM) install --frozen-lockfile
	$(MAKE) generate-drift
	$(MAKE) migration-drift
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) build

build:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) build ./cmd/server
	$(GO_RUN_ENV) $(GO) build ./cmd/migrate
	$(PNPM) --dir apps/web build
