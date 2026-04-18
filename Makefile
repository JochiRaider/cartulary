.PHONY: bootstrap bootstrap-node-runtime frontend-toolchain playwright-install db-up db-reset dev generate generate-drift migration-drift deployable-shape phase2-map-check backend-unit backend-integration frontend-unit browser-e2e test e2e lint check ci build

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
PNPM ?= $(shell if command -v pnpm >/dev/null 2>&1; then command -v pnpm; elif [ -x "$$HOME/.local/share/pnpm/pnpm" ]; then printf "$$HOME/.local/share/pnpm/pnpm"; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
NODE_RUNTIME_DIR ?= $(CURDIR)/tmp/node-runtime
NODE_BIN ?= $(NODE_RUNTIME_DIR)/bin/node
PNPM_RUN_ENV := PATH=$(NODE_RUNTIME_DIR)/bin:$$PATH

SQLC_TOOL := github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
GOOSE_TOOL := github.com/pressly/goose/v3/cmd/goose@v3.27.0
TESTCONTAINERS_GO_VERSION := v0.42.0

bootstrap-node-runtime:
	mkdir -p $(NODE_RUNTIME_DIR)
	@if [ -x "$(NODE_BIN)" ] && [ "$$($(NODE_BIN) -v)" = "v$(NODE_VERSION)" ]; then \
		: ; \
	else \
		archive="/tmp/node-v$(NODE_VERSION)-linux-x64.tar.xz"; \
		curl -fsSLo "$$archive" "https://nodejs.org/dist/v$(NODE_VERSION)/node-v$(NODE_VERSION)-linux-x64.tar.xz"; \
		rm -rf "$(NODE_RUNTIME_DIR)"; \
		mkdir -p "$(NODE_RUNTIME_DIR)"; \
		tar -xJf "$$archive" -C "$(NODE_RUNTIME_DIR)" --strip-components=1; \
	fi

frontend-toolchain: bootstrap-node-runtime
	@if [ -z "$(PNPM)" ]; then \
		echo "pnpm $(PNPM_VERSION) is required but was not found" >&2; \
		exit 1; \
	fi
	@if [ "$$($(PNPM_RUN_ENV) $(PNPM) --version)" != "$(PNPM_VERSION)" ]; then \
		echo "pnpm version mismatch: expected $(PNPM_VERSION)" >&2; \
		exit 1; \
	fi

bootstrap: frontend-toolchain
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) install $(SQLC_TOOL)
	$(GO_RUN_ENV) $(GO) install $(GOOSE_TOOL)
	$(PNPM_RUN_ENV) $(PNPM) install
	$(MAKE) --no-print-directory playwright-install

playwright-install: frontend-toolchain
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web exec playwright install chromium

db-up:
	docker compose -f docker-compose.dev.yml up -d postgres minio

db-reset:
	docker compose -f docker-compose.dev.yml up -d postgres
	docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS cartulary;"
	docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "CREATE DATABASE cartulary;"
	CARTULARY_CONFIG_FILE=$(CONFIG_FILE) $(GO_RUN_ENV) $(GO) run ./cmd/migrate up

dev: frontend-toolchain
	CARTULARY_CONFIG_FILE=$(CONFIG_FILE) CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH=$(CURDIR)/configs/dev/bootstrap-admin.json $(GO_RUN_ENV) $(GO) run ./cmd/server & \
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web dev & \
	wait

generate:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) run $(SQLC_TOOL) generate
	$(GO_RUN_ENV) $(GO) run ./tools/contractgen

# Codegen drift is distinct from migration drift.
generate-drift:
	./scripts/check-generate-drift.sh

# Migration drift covers schema-affecting changes not represented in /db/migrations
# or migrations that fail to apply cleanly in CI.
migration-drift:
	GO=$(GO) CONFIG_FILE=$(CONFIG_FILE) GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR) ./scripts/check-migrations.sh

deployable-shape:
	./scripts/ci/check-deployable-shape.sh

phase2-map-check: frontend-toolchain
	$(PNPM_RUN_ENV) $(NODE_BIN) ./scripts/check-phase2-map.mjs

test: frontend-toolchain
	@echo "== backend-unit =="
	$(MAKE) --no-print-directory backend-unit
	@echo "== backend-integration =="
	$(MAKE) --no-print-directory backend-integration
	@echo "== frontend-unit =="
	$(MAKE) --no-print-directory frontend-unit

backend-unit: frontend-toolchain
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) test ./internal/platform/... ./internal/testutil/configtest
	$(GO_RUN_ENV) $(GO) test ./internal/app ./internal/modules/auth ./internal/modules/incidents ./internal/modules/timeline -run '^(TestPhase0_.*_U_0_|TestPhase1_.*_U_1_|TestPhase2_.*_U_2_|TestPhase3_.*_U_3_)'

backend-integration: frontend-toolchain
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) test ./internal/testutil/httptestx ./internal/testutil/pgtest ./internal/testutil/s3test ./internal/testutil/wstest
	$(GO_RUN_ENV) $(GO) test ./internal/app ./internal/modules/auth ./internal/modules/incidents ./internal/modules/timeline -run '^(TestPhase0_.*_I_0_|TestPhase1_.*_I_1_|TestPhase2_.*_I_2_|TestPhase3_.*_I_3_)'
	$(GO_RUN_ENV) $(GO) test ./cmd/server -run '^(TestPhase1_.*_ProcessSmoke)$$'

frontend-unit: frontend-toolchain
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web test --run --passWithNoTests

e2e: frontend-toolchain
	@echo "== browser-e2e =="
	$(MAKE) --no-print-directory browser-e2e

browser-e2e: frontend-toolchain
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web test:e2e

lint: frontend-toolchain
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) vet ./...
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web exec biome check src
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web typecheck

check: frontend-toolchain
	$(PNPM_RUN_ENV) $(PNPM) install --frozen-lockfile
	@echo "== generate-drift =="
	$(MAKE) --no-print-directory generate-drift
	@echo "== migration-drift =="
	$(MAKE) --no-print-directory migration-drift
	@echo "== phase2-map-check =="
	$(MAKE) --no-print-directory phase2-map-check
	@echo "== lint =="
	$(MAKE) --no-print-directory lint
	@echo "== backend-unit =="
	$(MAKE) --no-print-directory backend-unit
	@echo "== backend-integration =="
	$(MAKE) --no-print-directory backend-integration
	@echo "== frontend-unit =="
	$(MAKE) --no-print-directory frontend-unit
	@echo "== browser-e2e =="
	$(MAKE) --no-print-directory browser-e2e
	@echo "== build =="
	$(MAKE) --no-print-directory build
	@echo "== deployable-shape =="
	$(MAKE) --no-print-directory deployable-shape

ci:
	./scripts/ci/verify.sh

build: frontend-toolchain
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_RUN_ENV) $(GO) build ./cmd/server
	$(GO_RUN_ENV) $(GO) build ./cmd/migrate
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web build
