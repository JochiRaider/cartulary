SHELL := /bin/bash

.PHONY: bootstrap bootstrap-node-runtime frontend-toolchain frontend-install frontend-install-ci playwright-install db-up db-reset dev generate generate-drift migration-drift deployable-shape deployable-shape-verify phase-map-check phase-test-name-check browser-e2e-task-surface-check backend-task-surface-check service-backed-unit-check run-phase-smoke backend-unit backend-store backend-integration backend-integration-support backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke frontend-unit browser-e2e browser-e2e-webserver-backed browser-e2e-functional browser-e2e-support browser-e2e-stateful browser-e2e-measurement browser-e2e-visual test-fast test-fast-service-backed-lane-a test-fast-service-backed-lane-b test e2e lint lint-go lint-biome lint-typecheck check check-preflight check-heavy check-service-backed check-service-backed-lane-a check-service-backed-lane-b check-isolated ci build build-server build-migrate build-web

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
PNPM ?= $(shell if command -v pnpm >/dev/null 2>&1; then command -v pnpm; elif [ -x "$$HOME/.local/share/pnpm/pnpm" ]; then printf "$$HOME/.local/share/pnpm/pnpm"; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
CHECK_JOBS ?= 4
SERVICE_BACKED_JOBS ?= 2
BACKEND_STORE_GO_TEST_P ?= 2
BACKEND_INTEGRATION_GO_TEST_P ?= 2
GO_TEST_SERVICE_PACKAGE_PARALLELISM ?= 1
PLAYWRIGHT_WORKERS ?= 2
VITEST_MAX_WORKERS ?= 2
NODE_RUNTIME_DIR ?= $(CURDIR)/tmp/node-runtime
NODE_BIN ?= $(NODE_RUNTIME_DIR)/bin/node
SERVER_BIN ?= $(CURDIR)/server
MIGRATE_BIN ?= $(CURDIR)/migrate
TOOLBIN_DIR ?= $(CURDIR)/tmp/toolbin
SQLC_BIN ?= $(TOOLBIN_DIR)/sqlc-v1.30.0
GOOSE_BIN ?= $(TOOLBIN_DIR)/goose-v3.27.0
TEST_SERVICES_BIN ?= $(TOOLBIN_DIR)/cartulary-test-services
FRONTEND_INSTALL_STAMP ?= $(CURDIR)/tmp/frontend-install/node-v$(NODE_VERSION)-pnpm-v$(PNPM_VERSION).stamp
PLAYWRIGHT_INSTALL_STAMP ?= $(CURDIR)/tmp/playwright/chromium.stamp
FRONTEND_TOOLCHAIN_STAMP ?= $(CURDIR)/tmp/frontend-toolchain/node-v$(NODE_VERSION)-pnpm-v$(PNPM_VERSION).stamp
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
RUN_FRONTEND_BIOME_SCRIPT := $(CURDIR)/scripts/run-frontend-biome.sh
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

define resolve_service_go_test_p
$(if $(filter environment environment override command line override,$(origin $(1))),$($(1)),$(if $(filter environment environment override command line override,$(origin GO_TEST_SERVICE_PACKAGE_PARALLELISM)),$(GO_TEST_SERVICE_PACKAGE_PARALLELISM),$($(1))))
endef

EFFECTIVE_BACKEND_STORE_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_STORE_GO_TEST_P)
EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_INTEGRATION_GO_TEST_P)

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

FRONTEND_INSTALL_INPUTS := package.json pnpm-lock.yaml pnpm-workspace.yaml apps/web/package.json $(wildcard packages/*/package.json)
SERVER_BUILD_INPUTS := go.mod go.sum $(shell rg --files cmd/server internal/app internal/modules internal/platform contracts 2>/dev/null)
MIGRATE_BUILD_INPUTS := go.mod go.sum $(shell rg --files cmd/migrate internal/app internal/platform db/migrations 2>/dev/null)
WEB_BUILD_INPUTS := package.json pnpm-lock.yaml pnpm-workspace.yaml $(shell rg --files apps/web packages 2>/dev/null)
TEST_SERVICES_BUILD_INPUTS := go.mod go.sum $(shell rg --files tools/testservices internal/testutil/pgtest internal/testutil/s3test internal/testutil/suiteservices internal/platform/postgres 2>/dev/null)

$(NODE_BIN):
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

bootstrap-node-runtime: $(NODE_BIN)

$(FRONTEND_TOOLCHAIN_STAMP): $(NODE_BIN)
	$(Q)mkdir -p $(dir $(FRONTEND_TOOLCHAIN_STAMP))
	$(Q)if [ -z "$(PNPM)" ]; then \
		echo "pnpm $(PNPM_VERSION) is required but was not found" >&2; \
		exit 1; \
	fi
	$(Q)if [ "$$($(PNPM_RUN_ENV) $(PNPM) --version)" != "$(PNPM_VERSION)" ]; then \
		echo "pnpm version mismatch: expected $(PNPM_VERSION)" >&2; \
		exit 1; \
	fi
	$(Q)printf 'node=%s\npnpm=%s\n' "$(NODE_VERSION)" "$(PNPM_VERSION)" > $(FRONTEND_TOOLCHAIN_STAMP)

frontend-toolchain: $(FRONTEND_TOOLCHAIN_STAMP)

$(FRONTEND_INSTALL_STAMP): $(FRONTEND_INSTALL_INPUTS) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(FRONTEND_INSTALL_STAMP))
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "frontend install" -- $(PNPM_ENV) $(PNPM) install $(PNPM_INSTALL_FLAGS)
	$(Q)printf 'node=%s\npnpm=%s\n' "$(NODE_VERSION)" "$(PNPM_VERSION)" > $(FRONTEND_INSTALL_STAMP)

frontend-install: $(FRONTEND_INSTALL_STAMP)

frontend-install-ci: $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(FRONTEND_INSTALL_STAMP))
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "check frontend install" -- $(PNPM_ENV) $(PNPM) install --frozen-lockfile $(PNPM_INSTALL_FLAGS)
	$(Q)printf 'node=%s\npnpm=%s\n' "$(NODE_VERSION)" "$(PNPM_VERSION)" > $(FRONTEND_INSTALL_STAMP)

$(SQLC_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/sqlc $(SQLC_BIN)
	$(RUN_PHASE) "bootstrap sqlc tool" -- env GOBIN=$(TOOLBIN_DIR) $(GO_RUN_ENV) $(GO) install $(SQLC_TOOL)
	$(Q)mv $(TOOLBIN_DIR)/sqlc $(SQLC_BIN)

$(GOOSE_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/goose $(GOOSE_BIN)
	$(RUN_PHASE) "bootstrap goose tool" -- env GOBIN=$(TOOLBIN_DIR) $(GO_RUN_ENV) $(GO) install $(GOOSE_TOOL)
	$(Q)mv $(TOOLBIN_DIR)/goose $(GOOSE_BIN)

$(TEST_SERVICES_BIN): $(TEST_SERVICES_BUILD_INPUTS)
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build testservices" -- $(GO_ENV) $(GO) build -o $(TEST_SERVICES_BIN) ./tools/testservices

$(PLAYWRIGHT_INSTALL_STAMP): $(FRONTEND_INSTALL_STAMP) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(PLAYWRIGHT_INSTALL_STAMP))
	$(RUN_PHASE) "playwright-install" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec playwright install chromium
	$(Q)printf 'node=%s\npnpm=%s\n' "$(NODE_VERSION)" "$(PNPM_VERSION)" > $(PLAYWRIGHT_INSTALL_STAMP)

bootstrap: $(SQLC_BIN) $(GOOSE_BIN) frontend-install playwright-install
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)

playwright-install: $(PLAYWRIGHT_INSTALL_STAMP)

db-up:
	$(Q)docker compose -f docker-compose.dev.yml up -d postgres minio

db-reset:
	$(Q)docker compose -f docker-compose.dev.yml up -d postgres
	$(Q)docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS cartulary;"
	$(Q)docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "CREATE DATABASE cartulary;"
	$(Q)env CARTULARY_CONFIG_FILE=$(CONFIG_FILE) $(GO_RUN_ENV) $(GO) run ./cmd/migrate up

dev: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)env CARTULARY_CONFIG_FILE=$(CONFIG_FILE) CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH=$(CURDIR)/configs/dev/bootstrap-admin.json $(GO_RUN_ENV) $(GO) run ./cmd/server & \
	$(PNPM_RUN_ENV) $(PNPM) --dir apps/web dev & \
	wait

generate: $(SQLC_BIN)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "generate sqlc" -- $(SQLC_BIN) generate
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

phase-map-check: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE) "phase-map-check" -- $(PNPM_ENV) env NODE_BIN=$(NODE_BIN) ./scripts/check-phase-maps.sh

run-phase-smoke: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE) "run-phase-smoke" -- bash -lc './scripts/test-run-phase.sh && ./scripts/test-run-go-target.sh && ./scripts/test-run-playwright-phase.sh && ./scripts/test-run-playwright-manifest-phase.sh && ./scripts/test-run-vitest-phase.sh && ./scripts/test-run-vitest-manifest-phase.sh && ./scripts/test-web-e2e-lifecycle.sh'

phase-test-name-check:
	$(RUN_PHASE) "phase-test-name-check" -- ./scripts/check-phase-test-names.sh

browser-e2e-task-surface-check:
	$(RUN_PHASE) "browser-e2e-task-surface-check" -- ./scripts/check-browser-e2e-task-surface.sh

backend-task-surface-check: $(NODE_BIN)
	$(RUN_PHASE) "backend-task-surface-check" -- env NODE_BIN=$(NODE_BIN) ./scripts/check-backend-task-surface.sh

service-backed-unit-check:
	$(RUN_PHASE) "service-backed-unit-check" -- ./scripts/check-service-backed-unit-tests.sh

test-fast: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) $(TEST_SERVICES_BIN)
	$(Q)$(MAKE) --no-print-directory --output-sync=target -j2 backend-unit frontend-unit
	$(Q)$(TEST_SERVICES_BIN) run -- $(MAKE) --no-print-directory --output-sync=target -j$(SERVICE_BACKED_JOBS) test-fast-service-backed-lane-a test-fast-service-backed-lane-b

test-fast-service-backed-lane-a:
	$(Q)$(MAKE) --no-print-directory backend-integration
	$(Q)$(MAKE) --no-print-directory backend-integration-support

test-fast-service-backed-lane-b: backend-store backend-process

test: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
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
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) test fail $$completed $$total test-fast backend-unit backend-store backend-integration backend-integration-support backend-process frontend-unit browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory browser-e2e; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) test fail $$completed $$total browser-e2e backend-unit backend-store backend-integration backend-integration-support backend-process frontend-unit browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi; \
		exit 1; \
	fi; \
	if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) test pass $$completed $$total - backend-unit backend-store backend-integration backend-integration-support backend-process frontend-unit browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi

backend-unit: export CARTULARY_TEST_TARGET := backend-unit
backend-unit: export CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION := phase1:unit:authoritative:backend_unit:./internal/platform/...

backend-unit: $(NODE_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) ./scripts/run-go-target.sh backend-unit

backend-store: export CARTULARY_TEST_TARGET := backend-store

backend-store: $(NODE_BIN) $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_PACKAGE_PARALLELISM=$(EFFECTIVE_BACKEND_STORE_GO_TEST_P) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-store

backend-integration: export CARTULARY_TEST_TARGET := backend-integration

backend-integration: $(NODE_BIN) $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_PACKAGE_PARALLELISM=$(EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-integration

backend-integration-support: export CARTULARY_TEST_TARGET := backend-integration-support

backend-integration-support: $(NODE_BIN) $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_PACKAGE_PARALLELISM=$(EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-integration-support

backend-process: export CARTULARY_TEST_TARGET := backend-process

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
backend-process: $(NODE_BIN) build-server $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) CARTULARY_SERVER_BIN=$(SERVER_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-process

phase0-process-e2e: export CARTULARY_TEST_TARGET := phase0-process-e2e

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
phase0-process-e2e: $(NODE_BIN) build-server $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) CARTULARY_SERVER_BIN=$(SERVER_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh phase0-process-e2e

phase1-process-smoke: export CARTULARY_TEST_TARGET := phase1-process-smoke

phase1-process-smoke: $(NODE_BIN) build-server $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) CARTULARY_SERVER_BIN=$(SERVER_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh phase1-process-smoke

phase2-process-smoke: export CARTULARY_TEST_TARGET := phase2-process-smoke

phase2-process-smoke: $(NODE_BIN) build-server $(TEST_SERVICES_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) CARTULARY_SERVER_BIN=$(SERVER_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh phase2-process-smoke

frontend-unit: export CARTULARY_TEST_TARGET := frontend-unit

frontend-unit: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)env PNPM=$(PNPM) NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) VITEST_FLAGS="$(VITEST_FLAGS)" VITEST_MAX_WORKERS=$(VITEST_MAX_WORKERS) ./scripts/run-frontend-unit.sh

e2e: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)$(MAKE) --no-print-directory browser-e2e

browser-e2e: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)$(MAKE) --no-print-directory browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual

browser-e2e-webserver-backed: export CARTULARY_TEST_TARGET := browser-e2e-webserver-backed

browser-e2e-webserver-backed: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(Q)env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-webserver-backed.sh
	$(TARGET_SUMMARY) browser-e2e-webserver-backed pass

browser-e2e-functional: export CARTULARY_TEST_TARGET := browser-e2e-functional

browser-e2e-functional: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(Q)env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-functional.sh
	$(TARGET_SUMMARY) browser-e2e-functional pass

browser-e2e-support: export CARTULARY_TEST_TARGET := browser-e2e-support

browser-e2e-support: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(RUN_PLAYWRIGHT_PHASE) "browser-e2e-support raw" -- env PLAYWRIGHT_WORKERS=$(PLAYWRIGHT_WORKERS) PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) $(PNPM) --dir apps/web exec playwright test e2e/phase2.support.spec.ts e2e/phase3.support.spec.ts
	$(TARGET_SUMMARY) browser-e2e-support pass

browser-e2e-stateful: export CARTULARY_TEST_TARGET := browser-e2e-stateful

# Browser evidence that mutates process-global backend state belongs here.
browser-e2e-stateful: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(Q)env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-stateful.sh $(PLAYWRIGHT_TEST_FLAGS)
	$(TARGET_SUMMARY) browser-e2e-stateful pass

browser-e2e-measurement: export CARTULARY_TEST_TARGET := browser-e2e-measurement

# Core 05-bound timing evidence is not parallel-safe with the heavy backend gate.
browser-e2e-measurement: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(Q)env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-measurement.sh
	$(TARGET_SUMMARY) browser-e2e-measurement pass

browser-e2e-visual: export CARTULARY_TEST_TARGET := browser-e2e-visual

browser-e2e-visual: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(Q)env PLAYWRIGHT_WORKERS=1 PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-visual.sh
	$(TARGET_SUMMARY) browser-e2e-visual pass

lint: lint-go lint-biome lint-typecheck

lint-go:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "lint go-vet" -- $(GO_ENV) $(GO) vet ./...

lint-biome: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)CARTULARY_PHASE_FAILURE_NOTE="run pnpm --dir apps/web format to apply the authoritative frontend Biome scope" CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT) "lint biome" -- bash $(RUN_FRONTEND_BIOME_SCRIPT) check $(BIOME_CHECK_FLAGS)

lint-typecheck: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "lint typecheck" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec tsc --noEmit $(TSC_FLAGS)

check-preflight: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)if [ "$(CI)" = "1" ]; then \
		$(MAKE) --no-print-directory frontend-install-ci; \
	else \
		$(MAKE) --no-print-directory frontend-install; \
	fi
	$(Q)$(MAKE) --no-print-directory lint-biome
	$(Q)$(MAKE) --no-print-directory run-phase-smoke
	$(Q)$(MAKE) --no-print-directory phase-test-name-check
	$(Q)$(MAKE) --no-print-directory browser-e2e-task-surface-check
	$(Q)$(MAKE) --no-print-directory backend-task-surface-check
	$(Q)$(MAKE) --no-print-directory service-backed-unit-check
	$(Q)$(MAKE) --no-print-directory generate-drift
	$(Q)$(MAKE) --no-print-directory phase-map-check

# Keep only parallel-safe work here. Service-backed Go phases and owned-stack
# browser suites run after this block under serialized orchestration.
check-heavy: migration-drift lint-go lint-typecheck backend-unit frontend-unit deployable-shape-verify

check-service-backed: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) $(TEST_SERVICES_BIN)
	$(Q)$(TEST_SERVICES_BIN) run -- $(MAKE) --no-print-directory --output-sync=target -j$(SERVICE_BACKED_JOBS) check-service-backed-lane-a check-service-backed-lane-b

check-service-backed-lane-a:
	$(Q)$(MAKE) --no-print-directory backend-integration
	$(Q)$(MAKE) --no-print-directory backend-integration-support

check-service-backed-lane-b: backend-store backend-process
	$(Q)$(MAKE) --no-print-directory browser-e2e-webserver-backed

check-isolated: browser-e2e-stateful browser-e2e-measurement browser-e2e-visual

check: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
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
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-preflight backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory --output-sync=target -j$(CHECK_JOBS) check-heavy; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-heavy backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory check-service-backed; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-service-backed backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi; \
		exit 1; \
	fi; \
	if $(MAKE) --no-print-directory check-isolated; then \
		completed=$$((completed + 1)); \
	else \
		if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check fail $$completed $$total check-isolated backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi; \
		exit 1; \
	fi; \
	if [ "$$dry_run" -ne 1 ]; then $(RUN_SUMMARY_CMD) check pass $$completed $$total - backend-unit frontend-unit backend-store backend-integration backend-integration-support backend-process browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; fi

ci:
	$(Q)./scripts/ci/verify.sh

build: build-server build-migrate build-web

$(SERVER_BIN): $(SERVER_BUILD_INPUTS)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build server" -- $(GO_ENV) $(GO) build -o $(SERVER_BIN) ./cmd/server

build-server: $(SERVER_BIN)

$(MIGRATE_BIN): $(MIGRATE_BUILD_INPUTS)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build migrate" -- $(GO_ENV) $(GO) build -o $(MIGRATE_BIN) ./cmd/migrate

build-migrate: $(MIGRATE_BIN)

$(CURDIR)/apps/web/dist/index.html: $(WEB_BUILD_INPUTS) $(FRONTEND_INSTALL_STAMP) | $(NODE_BIN)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "build web" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vite build $(VITE_BUILD_FLAGS)

build-web: $(CURDIR)/apps/web/dist/index.html
