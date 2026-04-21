SHELL := /bin/bash

.PHONY: bootstrap bootstrap-node-runtime frontend-toolchain playwright-install db-up db-reset dev generate generate-drift migration-drift deployable-shape deployable-shape-verify phase-map-check phase-test-name-check browser-e2e-task-surface-check backend-task-surface-check service-backed-unit-check run-phase-smoke backend-unit backend-store backend-integration backend-integration-support backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke frontend-unit browser-e2e browser-e2e-webserver-backed browser-e2e-functional browser-e2e-support browser-e2e-stateful browser-e2e-measurement test-fast test e2e lint lint-go lint-biome lint-typecheck check check-preflight check-heavy check-service-backed check-isolated ci build build-server build-migrate build-web

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
PNPM ?= $(shell if command -v pnpm >/dev/null 2>&1; then command -v pnpm; elif [ -x "$$HOME/.local/share/pnpm/pnpm" ]; then printf "$$HOME/.local/share/pnpm/pnpm"; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
CHECK_JOBS ?= 4
GO_TEST_SERVICE_PACKAGE_PARALLELISM ?= 1
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
RUN_GO_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-go-manifest-phase.sh
RUN_PLAYWRIGHT_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-playwright-phase.sh
RUN_PLAYWRIGHT_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-playwright-manifest-phase.sh
RUN_VITEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-vitest-phase.sh
RUN_VITEST_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-vitest-manifest-phase.sh
TEST_OUTPUT_SCRIPT := $(CURDIR)/scripts/lib/test-output.sh
RUN_PHASE = $(Q)$(RUN_PHASE_SCRIPT)
RUN_GO_PHASE = $(Q)$(RUN_GO_PHASE_SCRIPT)
RUN_GO_MANIFEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_GO_MANIFEST_PHASE_SCRIPT)
RUN_PLAYWRIGHT_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_PLAYWRIGHT_PHASE_SCRIPT)
RUN_PLAYWRIGHT_MANIFEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_PLAYWRIGHT_MANIFEST_PHASE_SCRIPT)
RUN_VITEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_VITEST_PHASE_SCRIPT)
RUN_VITEST_MANIFEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_VITEST_MANIFEST_PHASE_SCRIPT)
RUN_PHASE_ALLOW_SUCCESS_LOG = $(Q)CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT)
TARGET_SUMMARY = $(Q)NODE_BIN=$(NODE_BIN) $(TEST_OUTPUT_SCRIPT) target-summary
RUN_SUMMARY = $(Q)NODE_BIN=$(NODE_BIN) $(TEST_OUTPUT_SCRIPT) run-summary
RUN_SUMMARY_CMD = NODE_BIN=$(NODE_BIN) $(TEST_OUTPUT_SCRIPT) run-summary

CARTULARY_OUTPUT_MODE ?= quiet
CARTULARY_TEST_RESULTS_DIR ?= $(CURDIR)/.cartulary/test-results
CARTULARY_TEST_RUN_ID ?= $(shell printf '%s-p%s' "$$(date -u +%Y%m%dT%H%M%SZ)" "$$$$")
export CARTULARY_OUTPUT_MODE VERBOSE CI_VERBOSE CARTULARY_TEST_RESULTS_DIR CARTULARY_TEST_RUN_ID CARTULARY_TEST_INVENTORY

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
VITEST_FLAGS := --silent=passed-only
VITEST_MANIFEST_FLAGS := --silent=passed-only
PLAYWRIGHT_TEST_FLAGS := --quiet
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
	$(RUN_PHASE) "run-phase-smoke" -- bash -lc './scripts/test-run-phase.sh && ./scripts/test-run-playwright-phase.sh && ./scripts/test-run-playwright-manifest-phase.sh && ./scripts/test-run-vitest-phase.sh && ./scripts/test-run-vitest-manifest-phase.sh && ./scripts/test-web-e2e-lifecycle.sh'

phase-test-name-check:
	$(RUN_PHASE) "phase-test-name-check" -- ./scripts/check-phase-test-names.sh

browser-e2e-task-surface-check:
	$(RUN_PHASE) "browser-e2e-task-surface-check" -- ./scripts/check-browser-e2e-task-surface.sh

backend-task-surface-check:
	$(RUN_PHASE) "backend-task-surface-check" -- env NODE_BIN=$(NODE_BIN) ./scripts/check-backend-task-surface.sh

service-backed-unit-check:
	$(RUN_PHASE) "service-backed-unit-check" -- ./scripts/check-service-backed-unit-tests.sh

test-fast: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory backend-unit
	$(Q)$(MAKE) --no-print-directory backend-store
	$(Q)$(MAKE) --no-print-directory backend-integration
	$(Q)$(MAKE) --no-print-directory backend-integration-support
	$(Q)$(MAKE) --no-print-directory backend-process
	$(Q)$(MAKE) --no-print-directory frontend-unit

test: frontend-toolchain
	$(Q)set -e; \
	dry_run=0; \
	case " $(MAKEFLAGS) " in \
		*" n"*|*" --just-print"*|*" --dry-run"*) dry_run=1 ;; \
	esac; \
	completed=0; \
	total=2; \
	if $(MAKE) --no-print-directory test-fast; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) test fail $$completed $$total test-fast backend-unit backend-store backend-integration backend-integration-support backend-process frontend-unit browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory browser-e2e; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) test fail $$completed $$total browser-e2e backend-unit backend-store backend-integration backend-integration-support backend-process frontend-unit browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi; \
		exit 1; \
	fi; \
	if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) test pass $$completed $$total - backend-unit backend-store backend-integration backend-integration-support backend-process frontend-unit browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi

backend-unit: export CARTULARY_TEST_TARGET := backend-unit
backend-unit: export CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION := phase1:unit:authoritative:backend_unit:./internal/platform/...

backend-unit: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_MANIFEST_PHASE) "backend-unit phase0 authoritative platform" phase0 unit authoritative backend_unit -- $(GO_ENV) $(GO) test ./internal/platform/...
	$(RUN_GO_MANIFEST_PHASE) "backend-unit phase1 authoritative platform" phase1 unit authoritative backend_unit -- $(GO_ENV) $(GO) test ./internal/platform/...
	$(RUN_GO_PHASE) "backend-unit configtest" '^Test' -- $(GO_ENV) $(GO) test ./internal/testutil/configtest
	$(RUN_GO_MANIFEST_PHASE) "backend-unit phase0 authoritative app" phase0 unit authoritative backend_unit -- $(GO_ENV) $(GO) test ./internal/app
	$(RUN_GO_MANIFEST_PHASE) "backend-unit phase1 authoritative auth" phase1 unit authoritative backend_unit -- $(GO_ENV) $(GO) test ./internal/modules/auth
	$(RUN_GO_PHASE) "backend-unit support phase1" '^(TestSupportPhase1_)' -- $(GO_ENV) $(GO) test ./internal/modules/auth
	$(RUN_GO_PHASE) "backend-unit phase4 app" '^(TestPhase4_.*_U_4_0[89])' -- $(GO_ENV) $(GO) test ./internal/app ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline
	$(RUN_GO_MANIFEST_PHASE) "backend-unit phase2 authoritative" phase2 unit authoritative backend_unit -- $(GO_ENV) $(GO) test ./internal/modules/incidents
	$(RUN_GO_MANIFEST_PHASE) "backend-unit phase3 authoritative" phase3 unit authoritative backend_unit -- $(GO_ENV) $(GO) test ./internal/modules/timeline
	$(TARGET_SUMMARY) backend-unit pass

backend-store: export CARTULARY_TEST_TARGET := backend-store

backend-store: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "backend-store" '^(TestPhase4_.*_U_4_0[1-7])' -- $(GO_ENV) $(GO) test -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM) ./internal/modules/entities ./internal/modules/timeline
	$(RUN_GO_MANIFEST_PHASE) "backend-store phase2 authoritative" phase2 unit authoritative backend_store -- $(GO_ENV) $(GO) test ./internal/modules/incidents -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM)
	$(RUN_GO_MANIFEST_PHASE) "backend-store phase3 authoritative" phase3 unit authoritative backend_store -- $(GO_ENV) $(GO) test ./internal/modules/timeline -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM)
	$(TARGET_SUMMARY) backend-store pass

backend-integration: export CARTULARY_TEST_TARGET := backend-integration

backend-integration: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "backend-integration testutil" '^Test' -- $(GO_ENV) $(GO) test -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM) ./internal/testutil/httptestx ./internal/testutil/pgtest ./internal/testutil/s3test ./internal/testutil/wstest
	$(RUN_GO_MANIFEST_PHASE) "backend-integration phase0 authoritative" phase0 integration authoritative backend_integration -- $(GO_ENV) $(GO) test ./internal/platform/... ./internal/app -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM)
	$(RUN_GO_MANIFEST_PHASE) "backend-integration phase1 authoritative" phase1 integration authoritative backend_integration -- $(GO_ENV) $(GO) test ./internal/modules/auth -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM)
	$(RUN_GO_PHASE) "backend-integration phase4" '^(TestPhase4_.*_I_4_)' -- $(GO_ENV) $(GO) test -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM) ./internal/platform/... ./internal/app ./internal/modules/entities ./internal/modules/timeline
	$(RUN_GO_MANIFEST_PHASE) "backend-integration phase2 authoritative" phase2 integration authoritative backend_integration -- $(GO_ENV) $(GO) test ./internal/modules/incidents
	$(RUN_GO_MANIFEST_PHASE) "backend-integration phase3 authoritative" phase3 integration authoritative backend_integration -- $(GO_ENV) $(GO) test ./internal/modules/timeline -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM)
	$(TARGET_SUMMARY) backend-integration pass

backend-integration-support: export CARTULARY_TEST_TARGET := backend-integration-support

backend-integration-support: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "backend-integration support phase1" '^(TestSupportPhase1_)' -- $(GO_ENV) $(GO) test -p $(GO_TEST_SERVICE_PACKAGE_PARALLELISM) ./internal/modules/auth
	$(RUN_GO_PHASE) "backend-integration support phase2" '^(TestSupportPhase2_)' -- $(GO_ENV) $(GO) test ./internal/modules/incidents
	$(RUN_GO_PHASE) "backend-integration support phase3" '^(TestSupportPhase3_)' -- $(GO_ENV) $(GO) test ./internal/modules/timeline
	$(TARGET_SUMMARY) backend-integration-support pass

backend-process: export CARTULARY_TEST_TARGET := backend-process

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
backend-process: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_MANIFEST_PHASE) "backend-process phase0 authoritative" phase0 e2e authoritative backend_process -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4
	$(RUN_GO_PHASE) "backend-process phase1 smoke" '^(TestPhase1_.*_ProcessSmoke)$$' -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4
	$(TARGET_SUMMARY) backend-process pass

phase0-process-e2e: export CARTULARY_TEST_TARGET := phase0-process-e2e

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
phase0-process-e2e: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_MANIFEST_PHASE) "phase0-process-e2e" phase0 e2e authoritative backend_process -- $(GO_ENV) $(GO) test ./cmd/server
	$(TARGET_SUMMARY) phase0-process-e2e pass

phase1-process-smoke: export CARTULARY_TEST_TARGET := phase1-process-smoke

phase1-process-smoke: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "phase1-process-smoke" '^(TestPhase1_.*_ProcessSmoke)$$' -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4
	$(TARGET_SUMMARY) phase1-process-smoke pass

phase2-process-smoke: export CARTULARY_TEST_TARGET := phase2-process-smoke

phase2-process-smoke: frontend-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_GO_PHASE) "phase2-process-smoke" '^(TestPhase2_ProcessSmoke_)' -- $(GO_ENV) $(GO) test ./cmd/server -parallel 4
	$(TARGET_SUMMARY) phase2-process-smoke pass

frontend-unit: export CARTULARY_TEST_TARGET := frontend-unit

frontend-unit: frontend-toolchain
	$(RUN_VITEST_PHASE) "frontend-unit" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vitest run $(VITEST_FLAGS)
	$(RUN_VITEST_MANIFEST_PHASE) "frontend-unit phase2 authoritative" phase2 authoritative frontend_unit -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vitest run $(VITEST_MANIFEST_FLAGS)
	$(RUN_VITEST_MANIFEST_PHASE) "frontend-unit phase3 authoritative" phase3 authoritative frontend_unit -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vitest run $(VITEST_MANIFEST_FLAGS)
	$(TARGET_SUMMARY) frontend-unit pass

e2e: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory browser-e2e

browser-e2e: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory browser-e2e-webserver-backed
	$(Q)$(MAKE) --no-print-directory browser-e2e-stateful
	$(Q)$(MAKE) --no-print-directory browser-e2e-measurement

browser-e2e-webserver-backed: export CARTULARY_TEST_TARGET := browser-e2e-webserver-backed

browser-e2e-webserver-backed: frontend-toolchain build-server build-migrate
	$(Q)env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-webserver-backed.sh
	$(TARGET_SUMMARY) browser-e2e-webserver-backed pass

browser-e2e-functional: export CARTULARY_TEST_TARGET := browser-e2e-functional

browser-e2e-functional: frontend-toolchain build-server build-migrate
	$(Q)env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-functional.sh
	$(TARGET_SUMMARY) browser-e2e-functional pass

browser-e2e-support: export CARTULARY_TEST_TARGET := browser-e2e-support

browser-e2e-support: frontend-toolchain build-server build-migrate
	$(RUN_PLAYWRIGHT_PHASE) "browser-e2e-support phase2" -- env PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) $(PNPM) --dir apps/web exec playwright test e2e/phase2.support.spec.ts
	$(TARGET_SUMMARY) browser-e2e-support pass

browser-e2e-stateful: export CARTULARY_TEST_TARGET := browser-e2e-stateful

# Browser evidence that mutates process-global backend state belongs here.
browser-e2e-stateful: frontend-toolchain build-server build-migrate
	$(Q)env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-stateful.sh $(PLAYWRIGHT_TEST_FLAGS)
	$(TARGET_SUMMARY) browser-e2e-stateful pass

browser-e2e-measurement: export CARTULARY_TEST_TARGET := browser-e2e-measurement

# Core 05-bound timing evidence is not parallel-safe with the heavy backend gate.
browser-e2e-measurement: frontend-toolchain build-server build-migrate
	$(Q)env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-measurement.sh
	$(TARGET_SUMMARY) browser-e2e-measurement pass

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
	$(Q)$(MAKE) --no-print-directory browser-e2e-task-surface-check
	$(Q)$(MAKE) --no-print-directory backend-task-surface-check
	$(Q)$(MAKE) --no-print-directory service-backed-unit-check
	$(Q)$(MAKE) --no-print-directory generate-drift
	$(Q)$(MAKE) --no-print-directory phase-map-check

# Keep only parallel-safe work here. Service-backed Go phases and owned-stack
# browser suites run after this block under serialized orchestration.
check-heavy: migration-drift lint-go lint-biome lint-typecheck backend-unit frontend-unit deployable-shape-verify

check-service-backed: frontend-toolchain
	$(Q)$(MAKE) --no-print-directory backend-store
	$(Q)$(MAKE) --no-print-directory backend-integration
	$(Q)$(MAKE) --no-print-directory backend-integration-support
	$(Q)$(MAKE) --no-print-directory backend-process
	$(Q)$(MAKE) --no-print-directory browser-e2e-webserver-backed

check-isolated: browser-e2e-stateful browser-e2e-measurement

check: frontend-toolchain
	$(Q)set -e; \
	dry_run=0; \
	case " $(MAKEFLAGS) " in \
		*" n"*|*" --just-print"*|*" --dry-run"*) dry_run=1 ;; \
	esac; \
	completed=0; \
	total=4; \
	if $(MAKE) --no-print-directory check-preflight; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-preflight backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory --output-sync=target -j$(CHECK_JOBS) check-heavy; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-heavy backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory check-service-backed; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-service-backed backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory check-isolated; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-isolated backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi; \
		exit 1; \
	fi; \
	if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check pass $$completed $$total - backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement; fi

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
