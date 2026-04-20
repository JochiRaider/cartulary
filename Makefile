SHELL := /bin/bash

.PHONY: bootstrap bootstrap-node-runtime frontend-toolchain playwright-install db-up db-reset dev generate generate-drift migration-drift deployable-shape deployable-shape-verify phase-map-check phase-test-name-check run-phase-smoke backend-unit backend-integration backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke frontend-unit browser-e2e browser-e2e-functional browser-e2e-stateful browser-e2e-measurement test-fast test e2e lint lint-go lint-biome lint-typecheck check check-preflight check-heavy check-isolated ci build build-server build-migrate build-web

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
PNPM ?= $(shell if command -v pnpm >/dev/null 2>&1; then command -v pnpm; elif [ -x "$$HOME/.local/share/pnpm/pnpm" ]; then printf "$$HOME/.local/share/pnpm/pnpm"; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
CHECK_JOBS ?= 4
PLAYWRIGHT_WORKERS ?= 2
NODE_RUNTIME_DIR ?= $(CURDIR)/tmp/node-runtime
NODE_BIN ?= $(NODE_RUNTIME_DIR)/bin/node
SERVER_BIN ?= $(CURDIR)/server
MIGRATE_BIN ?= $(CURDIR)/migrate
PNPM_RUN_ENV := PATH=$(NODE_RUNTIME_DIR)/bin:$$PATH
GO_ENV := env $(GO_RUN_ENV)
PNPM_ENV := env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH"
Q := @
RUN_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-phase.sh
RUN_GO_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-go-phase.sh
RUN_PHASE = $(Q)$(RUN_PHASE_SCRIPT)
RUN_GO_PHASE = $(Q)$(RUN_GO_PHASE_SCRIPT)
RUN_PHASE_ALLOW_SUCCESS_LOG = $(Q)CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT)

CARTULARY_OUTPUT_MODE ?= quiet
export CARTULARY_OUTPUT_MODE VERBOSE CI_VERBOSE

ifeq ($(CI_VERBOSE),1)
EFFECTIVE_OUTPUT_MODE := normal
else ifeq ($(VERBOSE),1)
EFFECTIVE_OUTPUT_MODE := normal
else
EFFECTIVE_OUTPUT_MODE := $(CARTULARY_OUTPUT_MODE)
endif

ifeq ($(EFFECTIVE_OUTPUT_MODE),quiet)
PNPM_INSTALL_FLAGS := --reporter=append-only --loglevel=warn
BIOME_CHECK_FLAGS := --reporter=summary --diagnostic-level=warn
TSC_FLAGS := --pretty false
VITE_BUILD_FLAGS := --logLevel warn
VITEST_FLAGS := --reporter=dot --silent=passed-only
PLAYWRIGHT_TEST_FLAGS := --reporter=dot --quiet
endif

SQLC_TOOL := github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
GOOSE_TOOL := github.com/pressly/goose/v3/cmd/goose@v3.27.0
TESTCONTAINERS_GO_VERSION := v0.42.0

bootstrap-node-runtime:
	$(Q)mkdir -p $(NODE_RUNTIME_DIR)
	$(Q)if [ -x "$(NODE_BIN)" ] && [ "$$($(NODE_BIN) -v)" = "v$(NODE_VERSION)" ]; then \
		: ; \
	else \
		archive="/tmp/node-v$(NODE_VERSION)-linux-x64.tar.xz"; \
		curl -fsSLo "$$archive" "https://nodejs.org/dist/v$(NODE_VERSION)/node-v$(NODE_VERSION)-linux-x64.tar.xz"; \
		rm -rf "$(NODE_RUNTIME_DIR)"; \
		mkdir -p "$(NODE_RUNTIME_DIR)"; \
		tar -xJf "$$archive" -C "$(NODE_RUNTIME_DIR)" --strip-components=1; \
	fi

frontend-toolchain: bootstrap-node-runtime
	$(Q)if [ -z "$(PNPM)" ]; then \
		echo "pnpm $(PNPM_VERSION) is required but was not found" >&2; \
		exit 1; \
	fi
	$(Q)if [ "$$($(PNPM_RUN_ENV) $(PNPM) --version)" != "$(PNPM_VERSION)" ]; then \
		echo "pnpm version mismatch: expected $(PNPM_VERSION)" >&2; \
		exit 1; \
	fi

bootstrap: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "bootstrap sqlc tool" -- $(GO_ENV) $(GO) install $(SQLC_TOOL)
	$(RUN_PHASE) "bootstrap goose tool" -- $(GO_ENV) $(GO) install $(GOOSE_TOOL)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "bootstrap frontend install" -- $(PNPM_ENV) $(PNPM) install $(PNPM_INSTALL_FLAGS)
	$(Q)$(MAKE) --no-print-directory playwright-install

playwright-install: frontend-toolchain
	$(RUN_PHASE) "playwright-install" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec playwright install chromium

db-up:
	$(Q)docker compose -f docker-compose.dev.yml up -d postgres minio

db-reset:
	$(Q)docker compose -f docker-compose.dev.yml up -d postgres
	$(Q)docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS cartulary;"
	$(Q)docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "CREATE DATABASE cartulary;"
	$(Q)env CARTULARY_CONFIG_FILE=$(CONFIG_FILE) $(GO_RUN_ENV) $(GO) run ./cmd/migrate up

dev: frontend-toolchain
	$(Q)env CARTULARY_CONFIG_FILE=$(CONFIG_FILE) CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH=$(CURDIR)/configs/dev/bootstrap-admin.json $(GO_RUN_ENV) $(GO) run ./cmd/server & \
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web dev & \
	wait

generate:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "generate sqlc" -- $(GO_ENV) $(GO) run $(SQLC_TOOL) generate
	$(RUN_PHASE) "generate contracts" -- $(GO_ENV) $(GO) run ./tools/contractgen

# Codegen drift is distinct from migration drift.
generate-drift:
	$(RUN_PHASE) "generate-drift" -- ./scripts/check-generate-drift.sh

# Migration drift covers schema-affecting changes not represented in /db/migrations
# or migrations that fail to apply cleanly in CI.
migration-drift: build-migrate
	$(RUN_PHASE) "migration-drift" -- env GO=$(GO) CONFIG_FILE=$(CONFIG_FILE) GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/check-migrations.sh

deployable-shape: deployable-shape-verify

deployable-shape-verify: build-server build-migrate build-web
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "deployable-shape" -- ./scripts/ci/check-deployable-shape.sh

phase-map-check: frontend-toolchain
	$(RUN_PHASE) "phase-map-check" -- $(PNPM_ENV) env NODE_BIN=$(NODE_BIN) ./scripts/check-phase-maps.sh

run-phase-smoke:
	$(RUN_PHASE) "run-phase-smoke" -- ./scripts/test-run-phase.sh

phase-test-name-check:
	$(RUN_PHASE) "phase-test-name-check" -- ./scripts/check-phase-test-names.sh

test-fast: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory backend-unit
	$(Q)$(MAKE) --no-print-directory backend-integration
	$(Q)$(MAKE) --no-print-directory backend-process
	$(Q)$(MAKE) --no-print-directory frontend-unit

test: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory test-fast
	$(Q)$(MAKE) --no-print-directory browser-e2e

backend-unit: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "backend-unit platform" '^(TestPhase0_.*_U_0_|TestPhase1_.*_U_1_|TestPhase2_.*_U_2_|TestPhase3_.*_U_3_|TestPhase4_.*_U_4_)' -- $(GO_ENV) $(GO) test ./internal/platform/...
	$(RUN_PHASE) "backend-unit configtest" -- $(GO_ENV) $(GO) test ./internal/testutil/configtest
	$(RUN_GO_PHASE) "backend-unit phases" '^(TestPhase0_.*_U_0_|TestPhase1_.*_U_1_|TestPhase2_.*_U_2_|TestPhase3_.*_U_3_|TestPhase4_.*_U_4_)' -- $(GO_ENV) $(GO) test ./internal/app ./internal/modules/auth ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline

backend-integration: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "backend-integration testutil" -- $(GO_ENV) $(GO) test ./internal/testutil/httptestx ./internal/testutil/pgtest ./internal/testutil/s3test ./internal/testutil/wstest
	$(RUN_GO_PHASE) "backend-integration phases" '^(TestPhase0_.*_I_0_|TestPhase1_.*_I_1_|TestPhase2_.*_I_2_|TestPhase3_.*_I_3_|TestPhase4_.*_I_4_)' -- $(GO_ENV) $(GO) test ./internal/platform/... ./internal/app ./internal/modules/auth ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
backend-process: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "backend-process" '^(TestPhase0_.*_E_0_[0-9]+|TestPhase1_.*_ProcessSmoke|TestPhase2_ProcessSmoke_)$$' -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
phase0-process-e2e:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "phase0-process-e2e" '^(TestPhase0_.*_E_0_[0-9]+)$$' -- $(GO_ENV) $(GO) test ./cmd/server

phase1-process-smoke:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "phase1-process-smoke" '^(TestPhase1_.*_ProcessSmoke)$$' -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4

phase2-process-smoke:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "phase2-process-smoke" '^(TestPhase2_ProcessSmoke_)' -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4

frontend-unit: frontend-toolchain
	$(RUN_PHASE) "frontend-unit" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vitest run $(VITEST_FLAGS)

e2e: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory browser-e2e

browser-e2e: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory browser-e2e-functional
	$(Q)$(MAKE) --no-print-directory browser-e2e-stateful
	$(Q)$(MAKE) --no-print-directory browser-e2e-measurement

browser-e2e-functional: frontend-toolchain build-server build-migrate
	$(RUN_PHASE) "browser-e2e-functional" -- env PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) $(PNPM) --dir apps/web exec playwright test $(PLAYWRIGHT_TEST_FLAGS) e2e/phase1.spec.ts e2e/phase2.spec.ts e2e/phase3.spec.ts e2e/phase4.spec.ts

# Browser evidence that mutates process-global backend state belongs here.
browser-e2e-stateful: frontend-toolchain build-server build-migrate
	$(RUN_PHASE) "browser-e2e-stateful" -- env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) $(PNPM) --dir apps/web exec playwright test $(PLAYWRIGHT_TEST_FLAGS) e2e/phase1.clock.spec.ts

# Core 05-bound timing evidence is not parallel-safe with the heavy backend gate.
browser-e2e-measurement: frontend-toolchain build-server build-migrate
	$(RUN_PHASE) "browser-e2e-measurement" -- env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) $(PNPM) --dir apps/web exec playwright test $(PLAYWRIGHT_TEST_FLAGS) e2e/measurement/phase3_measurement.spec.ts

lint: lint-go lint-biome lint-typecheck

lint-go:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "lint go-vet" -- $(GO_ENV) $(GO) vet ./...

lint-biome: frontend-toolchain
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "lint biome" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec biome check src $(BIOME_CHECK_FLAGS)

lint-typecheck: frontend-toolchain
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "lint typecheck" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec tsc --noEmit $(TSC_FLAGS)

check-preflight: frontend-toolchain
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "check frontend install" -- $(PNPM_ENV) $(PNPM) install --frozen-lockfile $(PNPM_INSTALL_FLAGS)
	$(Q)$(MAKE) --no-print-directory run-phase-smoke
	$(Q)$(MAKE) --no-print-directory phase-test-name-check
	$(Q)$(MAKE) --no-print-directory generate-drift
	$(Q)$(MAKE) --no-print-directory phase-map-check

# Keep only parallel-safe work here; measurement-sensitive browser evidence runs after this block.
check-heavy: migration-drift lint-go lint-biome lint-typecheck backend-unit backend-integration backend-process frontend-unit browser-e2e-functional deployable-shape-verify

check-isolated: browser-e2e-stateful browser-e2e-measurement

check: check-preflight
	$(Q)$(MAKE) --no-print-directory --output-sync=target -j$(CHECK_JOBS) check-heavy
	$(Q)$(MAKE) --no-print-directory check-isolated

ci:
	$(Q)./scripts/ci/verify.sh

build: build-server build-migrate build-web

build-server:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build server" -- $(GO_ENV) $(GO) build ./cmd/server

build-migrate:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build migrate" -- $(GO_ENV) $(GO) build ./cmd/migrate

build-web: frontend-toolchain
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "build web" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vite build $(VITE_BUILD_FLAGS)
