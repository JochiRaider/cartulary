SHELL := /bin/bash
.DEFAULT_GOAL := help

include tools/task_surface.generated.mk

.SECONDEXPANSION:

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
CHECK_JOBS ?= 8
CHECK_IO_JOBS ?= 12
HARNESS_SMOKE_JOBS ?= 4
BACKEND_STORE_GO_TEST_P ?= 2
BACKEND_INTEGRATION_GO_TEST_P ?= 2
BACKEND_INTEGRATION_SHARD_JOBS ?= 4
GO_TEST_SERVICE_PACKAGE_PARALLELISM ?= 1
PLAYWRIGHT_WORKERS ?= 2
VITEST_MAX_WORKERS ?= 2
FIXTURE_THRESHOLD_MS ?= 30000
FIXTURE_TOP ?= 5
NODE_RUNTIME_DIR ?= $(CURDIR)/tmp/node-runtime
NODE_BIN ?= $(NODE_RUNTIME_DIR)/bin/node
PNPM ?= $(NODE_RUNTIME_DIR)/bin/pnpm
SERVER_BIN ?= $(CURDIR)/server
MIGRATE_BIN ?= $(CURDIR)/migrate
TOOLBIN_DIR ?= $(CURDIR)/tmp/toolbin
SQLC_BIN ?= $(TOOLBIN_DIR)/sqlc-v1.30.0
GOOSE_BIN ?= $(TOOLBIN_DIR)/goose-v3.27.0
TEST_SERVICES_BIN ?= $(TOOLBIN_DIR)/cartulary-test-services
SERVICE_BACKED_SCHEDULE_MANIFEST ?= $(CURDIR)/tools/service_backed_schedule_manifest.json
SERVICE_BACKED_SCHEDULE_PROFILE ?= $(CURDIR)/tools/service_backed_schedule_profiles.json
CHECK_SCHEDULE_MANIFEST ?= $(CURDIR)/tools/check_schedule_manifest.json
MINIO_BUCKET ?= cartulary
FRONTEND_INSTALL_STAMP ?= $(CURDIR)/tmp/frontend-install/node-v$(NODE_VERSION)-pnpm-v$(PNPM_VERSION).stamp
PLAYWRIGHT_INSTALL_STAMP ?= $(CURDIR)/tmp/playwright/chromium.stamp
FRONTEND_TOOLCHAIN_STAMP ?= $(CURDIR)/tmp/frontend-toolchain/node-v$(NODE_VERSION)-pnpm-v$(PNPM_VERSION).stamp
PNPM_RUN_ENV := PATH=$(NODE_RUNTIME_DIR)/bin:$$PATH COREPACK_HOME=$(NODE_RUNTIME_DIR)/corepack
GO_ENV := env $(GO_RUN_ENV)
PNPM_ENV := env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack"
BROWSER_E2E_OWNED_STACK_ENV := PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) CARTULARY_TEST_SERVICES_BIN=$(TEST_SERVICES_BIN) CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES=1
Q := @
comma := ,
RUN_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-phase.sh
RUN_GO_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-go-phase.sh
RUN_GO_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-go-manifest-phase.sh
RUN_PLAYWRIGHT_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-playwright-phase.sh
RUN_PLAYWRIGHT_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-playwright-manifest-phase.sh
RUN_VITEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-vitest-phase.sh
RUN_VITEST_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-vitest-manifest-phase.sh
RUN_FRONTEND_BIOME_SCRIPT := $(CURDIR)/scripts/run-frontend-biome.sh
TEST_OUTPUT_SCRIPT := $(CURDIR)/scripts/lib/test-output.sh
TASK_SURFACE_MANIFEST ?= $(CURDIR)/tools/task_surface_manifest.json
TASK_SURFACE_REPORT_ARGS ?=
RUN_MAKE_SEQUENCE_SCRIPT := $(CURDIR)/scripts/run-make-sequence.sh
RUN_HARNESS_SMOKE_SCRIPT := $(CURDIR)/scripts/run-harness-smoke.mjs
RUN_SERVICE_BACKED_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-service-backed-schedule.mjs
RUN_SERVICE_BACKED_SCHEDULE_TARGET_SCRIPT := $(CURDIR)/scripts/run-service-backed-schedule-target.sh
RUN_CHECK_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-check-schedule.mjs
BUILD_INPUTS_SCRIPT := $(CURDIR)/scripts/list-build-inputs.sh
DEV_SERVICES_SCRIPT := $(CURDIR)/scripts/dev-services.sh
RUN_PHASE = $(Q)$(RUN_PHASE_SCRIPT)
RUN_GO_PHASE = $(Q)$(RUN_GO_PHASE_SCRIPT)
RUN_GO_MANIFEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_GO_MANIFEST_PHASE_SCRIPT)
RUN_PLAYWRIGHT_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_PLAYWRIGHT_PHASE_SCRIPT)
RUN_PLAYWRIGHT_MANIFEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_PLAYWRIGHT_MANIFEST_PHASE_SCRIPT)
RUN_VITEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_VITEST_PHASE_SCRIPT)
RUN_VITEST_MANIFEST_PHASE = $(Q)NODE_BIN=$(NODE_BIN) $(RUN_VITEST_MANIFEST_PHASE_SCRIPT)
RUN_PHASE_ALLOW_SUCCESS_LOG = $(Q)CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT)
TARGET_SUMMARY = $(Q)NODE_BIN=$(NODE_BIN) TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" $(TEST_OUTPUT_SCRIPT) target-summary

define run_service_backed_schedule_target
$(Q)env MAKE="$(MAKE)" NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" TEST_SERVICES_BIN="$(TEST_SERVICES_BIN)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" RUN_SERVICE_BACKED_SCHEDULE_SCRIPT="$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT)" SERVICE_BACKED_SCHEDULE_MANIFEST="$(SERVICE_BACKED_SCHEDULE_MANIFEST)" $(RUN_SERVICE_BACKED_SCHEDULE_TARGET_SCRIPT) --target $(1) --phase-label "$(2)" --service-wrapper test-services
endef

define run_browser_batch_target
$(Q)env $(BROWSER_E2E_OWNED_STACK_ENV) TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" PLAYWRIGHT_WORKERS=$(2) $(if $(filter test-services,$(3)),$(TEST_SERVICES_BIN) run -- ,)./scripts/run-browser-e2e-target.sh $(1)
endef

define resolve_service_go_test_p
$(if $(filter environment environment override command line override,$(origin $(1))),$($(1)),$(if $(filter environment environment override command line override,$(origin GO_TEST_SERVICE_PACKAGE_PARALLELISM)),$(GO_TEST_SERVICE_PACKAGE_PARALLELISM),$($(1))))
endef

EFFECTIVE_BACKEND_STORE_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_STORE_GO_TEST_P)
EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_INTEGRATION_GO_TEST_P)

CARTULARY_OUTPUT_MODE ?= quiet
CARTULARY_TEST_RESULTS_DIR ?= $(CURDIR)/.cartulary/test-results
CARTULARY_TEST_RUN_ID ?= $(shell if [ -x /usr/bin/date ]; then now="$$(/usr/bin/date -u +%Y%m%dT%H%M%SZ)"; elif command -v date >/dev/null 2>&1; then now="$$(date -u +%Y%m%dT%H%M%SZ)"; else now="unknown-time"; fi; printf '%s-p%s' "$$now" "$$$$")
RELEASE_ARTIFACT_DIR ?= $(CURDIR)/.cartulary/release-artifacts
LICENSE_REPORT_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/license-report.json
SBOM_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/sbom.cdx.json
BENCHMARK_MANIFEST ?= $(CURDIR)/.cartulary/benchmark/benchmark_manifest.json
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
define discover_build_inputs
$(strip $(shell $(BUILD_INPUTS_SCRIPT) $(1)))$(if $(filter-out 0,$(.SHELLSTATUS)),$(error build input discovery failed for roots: $(1)),)
endef
SERVER_BUILD_INPUTS = go.mod go.sum $(call discover_build_inputs,cmd/server internal/app internal/modules internal/platform contracts)
MIGRATE_BUILD_INPUTS = go.mod go.sum $(call discover_build_inputs,cmd/migrate internal/app internal/platform db/migrations)
WEB_BUILD_INPUTS = package.json pnpm-lock.yaml pnpm-workspace.yaml $(call discover_build_inputs,apps/web packages)
TEST_SERVICES_BUILD_INPUTS = go.mod go.sum $(call discover_build_inputs,tools/testservices internal/testutil/pgtest internal/testutil/s3test internal/testutil/suiteservices internal/platform/postgres db/migrations)
EMBEDDED_WEB_ASSET_DIR := $(CURDIR)/internal/platform/httpapi/webassets/dist
EMBEDDED_WEB_ASSET_STAMP := $(CURDIR)/tmp/frontend-embed/web-assets.stamp
CLEAN_PATHS := $(SERVER_BIN) $(MIGRATE_BIN) $(CURDIR)/apps/web/dist $(EMBEDDED_WEB_ASSET_STAMP) $(CARTULARY_TEST_RESULTS_DIR) $(RELEASE_ARTIFACT_DIR) $(CURDIR)/apps/web/test-results $(CURDIR)/tmp/dev-stack $(CURDIR)/tmp/dev-stack-lifecycle-smoke.* $(CURDIR)/tmp/web-e2e-lifecycle-smoke.* $(CURDIR)/tmp/vitest-json-sample.*
DISTCLEAN_PATHS := $(CLEAN_PATHS) $(NODE_RUNTIME_DIR) $(TOOLBIN_DIR) $(CURDIR)/tmp/frontend-install $(CURDIR)/tmp/frontend-toolchain $(CURDIR)/tmp/playwright $(CURDIR)/tmp/frontend-embed $(CURDIR)/.cache $(CURDIR)/.pnpm-store $(CURDIR)/apps/web/.vite $(CURDIR)/playwright-report $(CURDIR)/apps/web/playwright-report $(CURDIR)/coverage $(CURDIR)/apps/web/coverage

define guarded_remove_paths
set -euo pipefail; \
repo="$(CURDIR)"; \
for path in $(1); do \
	if [ -z "$$path" ] || [ "$$path" = "/" ] || [ "$$path" = "." ]; then \
		echo "refusing unsafe cleanup path: '$$path'" >&2; \
		exit 1; \
	fi; \
	case "$$path" in \
		"$$repo"/*) ;; \
		*) echo "refusing cleanup path outside repository: $$path" >&2; exit 1 ;; \
	esac; \
	if [ -e "$$path" ] || [ -L "$$path" ]; then \
		printf 'removing %s\n' "$${path#$$repo/}"; \
		rm -rf -- "$$path"; \
	fi; \
done
endef

define clean_embedded_web_assets
set -euo pipefail; \
repo="$(CURDIR)"; \
dir="$(EMBEDDED_WEB_ASSET_DIR)"; \
if [ -z "$$dir" ] || [ "$$dir" = "/" ] || [ "$$dir" = "." ]; then \
	echo "refusing unsafe embedded asset path: '$$dir'" >&2; \
	exit 1; \
fi; \
case "$$dir" in \
	"$$repo"/*) ;; \
	*) echo "refusing embedded asset path outside repository: $$dir" >&2; exit 1 ;; \
esac; \
if [ -d "$$dir" ]; then \
	printf 'removing embedded web assets under %s, preserving .keep\n' "$${dir#$$repo/}"; \
	find "$$dir" -mindepth 1 -maxdepth 1 ! -name '.keep' -exec rm -rf -- {} +; \
fi
endef

help:
	$(Q)printf '%s\n' $(TASK_SURFACE_HELP_LINES)

help-all:
	$(Q)printf '%s\n' $(TASK_SURFACE_HELP_ALL_LINES)

doctor:
	$(Q)set -e; \
	fail=0; \
	print_missing() { printf 'missing %s: %s\n' "$$1" "$$2"; fail=1; }; \
	go_path="$(GO)"; \
	if [ -n "$$go_path" ] && [ ! -x "$$go_path" ] && command -v "$$go_path" >/dev/null 2>&1; then \
		go_path="$$(command -v "$$go_path")"; \
	fi; \
	if [ -n "$$go_path" ] && [ -x "$$go_path" ]; then \
		go_version_line="$$("$$go_path" version)"; \
		go_version="$${go_version_line#go version }"; \
		go_version="$${go_version%% *}"; \
		if [[ "$$go_version" == go1.26* ]]; then \
			printf 'ok go: %s %s\n' "$$go_path" "$$go_version"; \
		else \
			printf 'missing go: expected Go 1.26, found %s at %s\n' "$$go_version" "$$go_path"; \
			fail=1; \
		fi; \
	else \
		print_missing go "install Go 1.26 or set GO=/path/to/go"; \
	fi; \
	node_path=""; \
	if [ -x "$(NODE_BIN)" ]; then \
		node_path="$(NODE_BIN)"; \
	fi; \
	if [ -n "$$node_path" ]; then \
		node_version="$$("$$node_path" --version)"; \
		if [ "$$node_version" = "v$(NODE_VERSION)" ]; then \
			printf 'ok node: %s %s\n' "$$node_path" "$$node_version"; \
		else \
			printf 'missing node: expected v%s, found %s at %s\n' "$(NODE_VERSION)" "$$node_version" "$$node_path"; \
			fail=1; \
		fi; \
	else \
		print_missing node "run make bootstrap-node-runtime"; \
	fi; \
	pnpm_path=""; \
	if [ -n "$(PNPM)" ] && [ -x "$(PNPM)" ]; then \
		pnpm_path="$(PNPM)"; \
	fi; \
	if [ -n "$$pnpm_path" ]; then \
		pnpm_version="$$(PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack" "$$pnpm_path" --version 2>/dev/null || true)"; \
		if [ "$$pnpm_version" = "$(PNPM_VERSION)" ]; then \
			printf 'ok pnpm: %s %s\n' "$$pnpm_path" "$$pnpm_version"; \
		else \
			printf 'missing pnpm: expected %s, found %s at %s\n' "$(PNPM_VERSION)" "$${pnpm_version:-unusable}" "$$pnpm_path"; \
			fail=1; \
		fi; \
	else \
		print_missing pnpm "run make frontend-toolchain"; \
	fi; \
	if command -v docker >/dev/null 2>&1; then \
		docker_path="$$(command -v docker)"; \
		compose_version="$$(docker compose version --short 2>/dev/null || docker compose version 2>/dev/null || true)"; \
		if [ -n "$$compose_version" ]; then \
			printf 'ok docker compose: %s %s\n' "$$docker_path" "$$compose_version"; \
		else \
			print_missing "docker compose" "install Docker with the compose plugin"; \
		fi; \
	else \
		print_missing "docker compose" "install Docker with the compose plugin"; \
	fi; \
	for tool in rg curl tar; do \
		if command -v "$$tool" >/dev/null 2>&1; then \
			tool_path="$$(command -v "$$tool")"; \
			tool_version="$$("$$tool" --version 2>/dev/null | head -n 1 || true)"; \
			printf 'ok %s: %s %s\n' "$$tool" "$$tool_path" "$$tool_version"; \
		else \
			print_missing "$$tool" "install $$tool and ensure it is on PATH"; \
		fi; \
	done; \
	if command -v ss >/dev/null 2>&1; then \
		ss_path="$$(command -v ss)"; \
		ss_version="$$(ss --version 2>&1 | head -n 1 || true)"; \
		printf 'ok ss: %s %s\n' "$$ss_path" "$$ss_version"; \
	else \
		printf 'optional missing ss: install iproute2 for local port diagnostics\n'; \
	fi; \
	exit "$$fail"

$(NODE_BIN): scripts/bootstrap-node-runtime.sh Makefile
	$(Q)NODE_VERSION="$(NODE_VERSION)" NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" ./scripts/bootstrap-node-runtime.sh

bootstrap-node-runtime: $(NODE_BIN)

$(FRONTEND_TOOLCHAIN_STAMP): $(NODE_BIN) Makefile
	$(Q)mkdir -p $(dir $(FRONTEND_TOOLCHAIN_STAMP))
	$(Q)$(PNPM_ENV) $(NODE_RUNTIME_DIR)/bin/corepack enable --install-directory "$(NODE_RUNTIME_DIR)/bin" pnpm
	$(Q)$(PNPM_ENV) $(NODE_RUNTIME_DIR)/bin/corepack prepare pnpm@$(PNPM_VERSION) --activate >/dev/null
	$(Q)node_version="$$($(NODE_BIN) --version)"; \
	if [ "$$node_version" != "v$(NODE_VERSION)" ]; then \
		echo "node version mismatch: expected v$(NODE_VERSION), got $$node_version at $(NODE_BIN)" >&2; \
		exit 1; \
	fi; \
	pnpm_version="$$($(PNPM_ENV) $(PNPM) --version)"; \
	if [ "$$pnpm_version" != "$(PNPM_VERSION)" ]; then \
		echo "pnpm version mismatch: expected $(PNPM_VERSION), got $$pnpm_version at $(PNPM)" >&2; \
		exit 1; \
	fi; \
	printf 'node_path=%s\nnode_version=%s\npnpm_path=%s\npnpm_version=%s\n' "$(NODE_BIN)" "$$node_version" "$(PNPM)" "$$pnpm_version" > $(FRONTEND_TOOLCHAIN_STAMP)

frontend-toolchain: $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)if [ "$${CARTULARY_FRONTEND_TOOLCHAIN_QUIET:-0}" != "1" ]; then cat $(FRONTEND_TOOLCHAIN_STAMP); fi

$(FRONTEND_INSTALL_STAMP): $(FRONTEND_INSTALL_INPUTS) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(FRONTEND_INSTALL_STAMP))
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "frontend install" -- $(PNPM_ENV) $(PNPM) install $(PNPM_INSTALL_FLAGS)
	$(Q)printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' "$(NODE_BIN)" "$(NODE_VERSION)" "$(PNPM)" "$(PNPM_VERSION)" > $(FRONTEND_INSTALL_STAMP)

frontend-install: $(FRONTEND_INSTALL_STAMP)

frontend-install-ci: $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(FRONTEND_INSTALL_STAMP))
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "check frontend install" -- $(PNPM_ENV) $(PNPM) install --frozen-lockfile $(PNPM_INSTALL_FLAGS)
	$(Q)printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' "$(NODE_BIN)" "$(NODE_VERSION)" "$(PNPM)" "$(PNPM_VERSION)" > $(FRONTEND_INSTALL_STAMP)

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

$(TEST_SERVICES_BIN): $$(TEST_SERVICES_BUILD_INPUTS)
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build testservices" -- $(GO_ENV) $(GO) build -o $(TEST_SERVICES_BIN) ./tools/testservices

$(PLAYWRIGHT_INSTALL_STAMP): $(FRONTEND_INSTALL_STAMP) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(PLAYWRIGHT_INSTALL_STAMP))
	$(RUN_PHASE) "playwright-install" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec playwright install chromium
	$(Q)printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' "$(NODE_BIN)" "$(NODE_VERSION)" "$(PNPM)" "$(PNPM_VERSION)" > $(PLAYWRIGHT_INSTALL_STAMP)

bootstrap: $(SQLC_BIN) $(GOOSE_BIN) frontend-install playwright-install
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)

playwright-install: $(PLAYWRIGHT_INSTALL_STAMP)

services-up:
	$(Q)docker compose -f docker-compose.dev.yml up -d postgres minio
	$(Q)$(MAKE) --no-print-directory services-wait

services-wait: postgres-wait minio-wait

postgres-wait:
	$(Q)$(DEV_SERVICES_SCRIPT) wait-postgres

minio-wait:
	$(Q)$(DEV_SERVICES_SCRIPT) wait-minio

minio-init: services-wait
	$(Q)MINIO_BUCKET="$(MINIO_BUCKET)" $(DEV_SERVICES_SCRIPT) init-minio

db-up:
	$(Q)$(MAKE) --no-print-directory services-up
	$(Q)$(MAKE) --no-print-directory minio-init

db-reset:
	$(Q)docker compose -f docker-compose.dev.yml up -d postgres
	$(Q)$(MAKE) --no-print-directory postgres-wait
	$(Q)printf '%s\n' 'db-reset: database reset only; MinIO/object storage is not reset.'
	$(Q)docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS cartulary;"
	$(Q)docker compose -f docker-compose.dev.yml exec -T postgres psql -U cartulary -d postgres -c "CREATE DATABASE cartulary;"
	$(Q)env CARTULARY_CONFIG_FILE=$(CONFIG_FILE) $(GO_RUN_ENV) $(GO) run ./cmd/migrate up

dev: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)env GO=$(GO) CONFIG_FILE=$(CONFIG_FILE) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) PNPM=$(PNPM) ./scripts/dev-stack.sh

codegen-toolchain: $(SQLC_BIN)

generate: codegen-toolchain
	$(Q)$(MAKE) --no-print-directory generate-artifacts

generate-artifacts:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "generate sqlc" -- $(SQLC_BIN) generate
	$(RUN_PHASE) "generate contracts" -- $(GO_ENV) $(GO) run ./tools/contractgen

# Codegen drift is distinct from migration drift.
generate-drift: codegen-toolchain
	$(RUN_PHASE) "generate-drift" -- ./scripts/check-generate-drift.sh

toolchain-drift: $(NODE_BIN)
	$(Q)$(NODE_BIN) ./scripts/check-toolchain-pins.mjs

# Migration drift covers schema-affecting changes not represented in /db/migrations
# or migrations that fail to apply cleanly in CI.
migration-drift: build-migrate
	$(RUN_PHASE) "migration-drift" -- env GO=$(GO) CONFIG_FILE=$(CONFIG_FILE) GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) ./scripts/check-migrations.sh

deployable-shape: build-server build-migrate
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "deployable-shape" -- ./scripts/ci/check-deployable-shape.sh

phase-map-check: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE) "phase-map-check" -- $(PNPM_ENV) env NODE_BIN=$(NODE_BIN) ./scripts/check-phase-maps.sh

phase-ledgers: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE) "phase-ledgers" -- $(PNPM_ENV) env NODE_BIN=$(NODE_BIN) $(NODE_BIN) ./scripts/render-phase-ledgers.mjs

phase-ledger-drift: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE) "phase-ledger-drift" -- $(PNPM_ENV) env NODE_BIN=$(NODE_BIN) $(NODE_BIN) ./scripts/check-phase-ledger-drift.mjs

phase-schedules: $(NODE_BIN)
	$(RUN_PHASE) "phase-schedules" -- $(NODE_BIN) ./scripts/render-service-backed-schedule-manifest.mjs --profile "$(SERVICE_BACKED_SCHEDULE_PROFILE)" --output "$(SERVICE_BACKED_SCHEDULE_MANIFEST)"

phase-schedule-drift: $(NODE_BIN)
	$(RUN_PHASE) "phase-schedule-drift" -- $(NODE_BIN) ./scripts/render-service-backed-schedule-manifest.mjs --check --profile "$(SERVICE_BACKED_SCHEDULE_PROFILE)" --output "$(SERVICE_BACKED_SCHEDULE_MANIFEST)"

benchmark-claim-check: $(NODE_BIN)
	$(RUN_PHASE) "benchmark-claim-check" -- $(NODE_BIN) ./scripts/check-benchmark-claim.mjs "$(BENCHMARK_MANIFEST)"

task-surface-report: $(NODE_BIN)
	$(Q)$(NODE_BIN) ./scripts/print-task-surface-report.mjs $(TASK_SURFACE_REPORT_ARGS)

task-surface-check: $(NODE_BIN)
	$(RUN_PHASE) "task-surface-check" -- $(NODE_BIN) ./scripts/print-task-surface-report.mjs --check

run-harness-smoke-fast: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" $(NODE_BIN) $(RUN_HARNESS_SMOKE_SCRIPT) --tier fast --jobs "$(HARNESS_SMOKE_JOBS)"

run-harness-smoke-extended: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" $(NODE_BIN) $(RUN_HARNESS_SMOKE_SCRIPT) --tier extended --jobs "$(HARNESS_SMOKE_JOBS)"

run-harness-smoke-full: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" $(NODE_BIN) $(RUN_HARNESS_SMOKE_SCRIPT) --tier full --jobs "$(HARNESS_SMOKE_JOBS)"

phase-test-name-check: $(NODE_BIN)
	$(RUN_PHASE) "phase-test-name-check" -- $(NODE_BIN) ./scripts/check-phase-test-names.mjs

browser-e2e-task-surface-check:
	$(RUN_PHASE) "browser-e2e-task-surface-check" -- ./scripts/check-browser-e2e-task-surface.sh

frontend-task-surface-check: $(NODE_BIN)
	$(RUN_PHASE) "frontend-task-surface-check" -- env NODE_BIN=$(NODE_BIN) ./scripts/check-frontend-task-surface.sh

backend-task-surface-check: $(NODE_BIN)
	$(RUN_PHASE) "backend-task-surface-check" -- env NODE_BIN=$(NODE_BIN) ./scripts/check-backend-task-surface.sh

service-backed-unit-check:
	$(RUN_PHASE) "service-backed-unit-check" -- ./scripts/check-service-backed-unit-tests.sh

test-service-images: $(TEST_SERVICES_BIN)
	$(RUN_PHASE) "warm test service images" -- $(TEST_SERVICES_BIN) warm-images

test-local: backend-unit frontend-typecheck frontend-unit

test-fast: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) $(TEST_SERVICES_BIN)
	$(Q)$(MAKE) --no-print-directory --output-sync=target -j3 test-local
	$(Q)$(MAKE) --no-print-directory test-fast-service-backed

test-fast-service-backed: export CARTULARY_TEST_TARGET := test-fast-service-backed

test-fast-service-backed: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server $(TEST_SERVICES_BIN) test-service-images
	$(call run_service_backed_schedule_target,test-fast-service-backed,test-fast service-backed)

test-service-backed: export CARTULARY_TEST_TARGET := test-service-backed

test-service-backed: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate $(TEST_SERVICES_BIN) test-service-images
	$(call run_service_backed_schedule_target,test-service-backed,test service-backed)

test: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)MAKE="$(MAKE)" NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" $(RUN_MAKE_SEQUENCE_SCRIPT) --label test --summary-profile test --parallel-step test-local:3 --step test-service-backed

backend-unit: export CARTULARY_TEST_TARGET := backend-unit
backend-unit: export CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION := phase1:unit:authoritative:backend_unit:./internal/platform/...

backend-unit: $(NODE_BIN)
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) ./scripts/run-go-target.sh backend-unit

target-plan:
	$(Q)node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/print-target-plan.mjs

target-plan-json:
	$(Q)node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/print-target-plan.mjs --json

fixture-report:
	$(Q)set -euo pipefail; \
	node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; \
	results_dir="$(if $(RESULTS_DIR),$(RESULTS_DIR),$(CARTULARY_TEST_RESULTS_DIR))"; \
	args=(--results-dir "$$results_dir" --threshold-ms "$(FIXTURE_THRESHOLD_MS)" --top "$(FIXTURE_TOP)"); \
	if [ -n "$(RUN_ID)" ]; then args+=(--run-id "$(RUN_ID)"); fi; \
	if [ -n "$(TARGET)" ]; then args+=(--target "$(TARGET)"); fi; \
	if [ "$(JSON)" = "1" ]; then args+=(--json); fi; \
	"$$node_cmd" ./scripts/print-fixture-report.mjs "$${args[@]}"

explain-run:
	$(Q)if [ -z "$(RESULTS_DIR)" ]; then echo "usage: make explain-run RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] [TARGET=<target>] [DETAIL=summary|children|logs]" >&2; exit 2; fi
	$(Q)set -euo pipefail; \
	node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; \
	args=(--results-dir "$(RESULTS_DIR)" --detail "$(if $(DETAIL),$(DETAIL),summary)"); \
	if [ -n "$(RUN_ID)" ]; then args+=(--run-id "$(RUN_ID)"); fi; \
	if [ -n "$(TARGET)" ]; then args+=(--target "$(TARGET)"); fi; \
	"$$node_cmd" ./scripts/print-explain-run.mjs "$${args[@]}"

explain-target:
	$(Q)if [ -z "$(TARGET)" ]; then echo "usage: make explain-target TARGET=<backend target>" >&2; exit 2; fi
	$(Q)node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/print-target-plan.mjs --target "$(TARGET)" $(if $(filter 0,$(DETAIL)),,--detail)

go-test-duration-baselines:
	$(Q)if [ -z "$(RESULTS_DIR)" ]; then echo "usage: make go-test-duration-baselines RESULTS_DIR=<successful test results dir>" >&2; exit 2; fi
	$(Q)node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/update-go-test-durations.mjs $(if $(filter 1,$(PRUNE_OBSERVED_PACKAGES)),--prune-observed-packages) $(if $(BASELINE_FILE),--baseline-file "$(BASELINE_FILE)") "$(RESULTS_DIR)"

go-test-duration-baseline-coverage:
	$(Q)node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/check-go-test-duration-baseline-coverage.mjs $(if $(BASELINE_FILE),--baseline-file "$(BASELINE_FILE)")

go-test-duration-baseline-drift:
	$(Q)results_dir="$(RESULTS_DIR)"; if [ -z "$$results_dir" ]; then results_dir="$(CARTULARY_TEST_RESULTS_DIR)/$(CARTULARY_TEST_RUN_ID)"; fi; node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/check-go-test-duration-baseline-drift.mjs $(if $(BASELINE_FILE),--baseline-file "$(BASELINE_FILE)") "$$results_dir"

browser-e2e-duration-baseline-drift:
	$(Q)results_dir="$(RESULTS_DIR)"; if [ -z "$$results_dir" ]; then results_dir="$(CARTULARY_TEST_RESULTS_DIR)/$(CARTULARY_TEST_RUN_ID)"; fi; node_cmd="$(NODE_BIN)"; if [ ! -x "$$node_cmd" ]; then node_cmd=node; fi; "$$node_cmd" ./scripts/lib/browser-shard-plan.mjs check-baseline-drift $(if $(BASELINE_FILE),--baseline-file "$(BASELINE_FILE)") "$$results_dir"

backend-store: export CARTULARY_TEST_TARGET := backend-store

backend-store: $(NODE_BIN) $(TEST_SERVICES_BIN) test-service-images
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_PACKAGE_PARALLELISM=$(EFFECTIVE_BACKEND_STORE_GO_TEST_P) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-store

backend-integration: export CARTULARY_TEST_TARGET := backend-integration

backend-integration: $(NODE_BIN) $(TEST_SERVICES_BIN) test-service-images
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_PACKAGE_PARALLELISM=$(EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) BACKEND_INTEGRATION_SHARD_JOBS=$(BACKEND_INTEGRATION_SHARD_JOBS) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-integration

backend-integration-support: export CARTULARY_TEST_TARGET := backend-integration-support

backend-integration-support: $(NODE_BIN) $(TEST_SERVICES_BIN) test-service-images
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) GO_TEST_PACKAGE_PARALLELISM=$(EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) BACKEND_INTEGRATION_SHARD_JOBS=$(BACKEND_INTEGRATION_SHARD_JOBS) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-integration-support

backend-process: export CARTULARY_TEST_TARGET := backend-process

# Phase 0 process evidence is part of the developer gate and must never be direct-run only.
backend-process: $(NODE_BIN) build-server $(TEST_SERVICES_BIN) test-service-images
	$(Q)env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) NODE_BIN=$(NODE_BIN) CARTULARY_SERVER_BIN=$(SERVER_BIN) GO_TEST_SERVICE_PACKAGE_PARALLELISM=$(GO_TEST_SERVICE_PACKAGE_PARALLELISM) $(TEST_SERVICES_BIN) run -- ./scripts/run-go-target.sh backend-process

frontend-unit: export CARTULARY_TEST_TARGET := frontend-unit

frontend-typecheck: export CARTULARY_TEST_TARGET := frontend-typecheck

frontend-typecheck: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "frontend typecheck" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec tsc --noEmit $(TSC_FLAGS)
	$(TARGET_SUMMARY) frontend-typecheck pass

frontend-unit: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)env PNPM=$(PNPM) NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) VITEST_FLAGS="$(VITEST_FLAGS)" VITEST_MAX_WORKERS=$(VITEST_MAX_WORKERS) ./scripts/run-frontend-unit.sh

browser-e2e: export CARTULARY_TEST_TARGET := browser-e2e

browser-e2e: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate $(TEST_SERVICES_BIN) test-service-images
	$(call run_browser_batch_target,isolated,1,test-services)

browser-e2e-webserver-backed: export CARTULARY_TEST_TARGET := browser-e2e-webserver-backed

browser-e2e-webserver-backed: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,webserver-backed,$(PLAYWRIGHT_WORKERS),direct)

browser-e2e-functional: export CARTULARY_TEST_TARGET := browser-e2e-functional

browser-e2e-functional: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,functional,$(PLAYWRIGHT_WORKERS),direct)

browser-e2e-support: export CARTULARY_TEST_TARGET := browser-e2e-support

browser-e2e-support: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,support,$(PLAYWRIGHT_WORKERS),direct)

browser-e2e-stateful: export CARTULARY_TEST_TARGET := browser-e2e-stateful

# Browser evidence that mutates process-global backend state belongs here.
browser-e2e-stateful: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,stateful,1,direct)

browser-e2e-resettable: export CARTULARY_TEST_TARGET := browser-e2e-resettable

browser-e2e-resettable: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,resettable,1,direct)

browser-e2e-measurement: export CARTULARY_TEST_TARGET := browser-e2e-measurement

# Ordinary implementation/regression measurement; not claim-bearing Core 05 publication evidence.
browser-e2e-measurement: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,measurement,1,direct)

browser-e2e-visual: export CARTULARY_TEST_TARGET := browser-e2e-visual

browser-e2e-visual: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate
	$(call run_browser_batch_target,visual,1,direct)

lint: lint-go lint-biome frontend-typecheck

lint-go:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "lint go-vet" -- bash ./scripts/run-go-vet.sh

format: format-frontend

format-frontend: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT) "format frontend" -- bash $(RUN_FRONTEND_BIOME_SCRIPT) format

lint-biome: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)CARTULARY_PHASE_FAILURE_NOTE="run make format to apply the authoritative frontend Biome scope" CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT) "lint biome" -- bash $(RUN_FRONTEND_BIOME_SCRIPT) check $(BIOME_CHECK_FLAGS)

check-setup-blockers: $(NODE_BIN)
	$(Q)$(MAKE) --no-print-directory toolchain-drift
	$(Q)$(MAKE) --no-print-directory codegen-toolchain
	$(Q)if [ "$(CI)" = "1" ]; then \
		$(MAKE) --no-print-directory frontend-install-ci; \
	else \
		$(MAKE) --no-print-directory frontend-install; \
	fi

check-static-validation:
	$(Q)$(MAKE) --no-print-directory lint-biome
	$(Q)$(MAKE) --no-print-directory phase-test-name-check
	$(Q)$(MAKE) --no-print-directory task-surface-check
	$(Q)$(MAKE) --no-print-directory browser-e2e-task-surface-check
	$(Q)$(MAKE) --no-print-directory frontend-task-surface-check
	$(Q)$(MAKE) --no-print-directory backend-task-surface-check
	$(Q)$(MAKE) --no-print-directory phase-map-check
	$(Q)$(MAKE) --no-print-directory go-test-duration-baseline-coverage
	$(Q)$(MAKE) --no-print-directory phase-ledger-drift
	$(Q)$(MAKE) --no-print-directory phase-schedule-drift
	$(Q)$(MAKE) --no-print-directory service-backed-unit-check
	$(Q)$(MAKE) --no-print-directory generate-drift

check-harness-smoke:
	$(Q)$(MAKE) --no-print-directory run-harness-smoke-fast
	$(TARGET_SUMMARY) check-harness-smoke pass --projection check-harness-smoke

# Build and service-image readiness gate the service-backed scheduler. Keep
# independent static and harness validation outside this prerequisite aggregate
# so service-backed work can start as soon as these artifacts are ready.
check-build-prereqs: build-server build-migrate test-service-images

check-local-product: migration-drift lint-go frontend-typecheck backend-unit deployable-shape

check-go-test-duration-baseline-drift:
	$(Q)$(MAKE) --no-print-directory go-test-duration-baseline-drift
	$(TARGET_SUMMARY) check-go-test-duration-baseline-drift pass

check-browser-e2e-duration-baseline-drift:
	$(Q)$(MAKE) --no-print-directory browser-e2e-duration-baseline-drift
	$(TARGET_SUMMARY) check-browser-e2e-duration-baseline-drift pass

check-frontend-unit: frontend-unit

check-meta-validation: check-static-validation check-harness-smoke

check-service-backed: export CARTULARY_TEST_TARGET := check-service-backed

check-service-backed: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) build-server build-migrate test-service-images
	$(call run_service_backed_schedule_target,check-service-backed,check service-backed)

check: $(NODE_BIN)
	$(Q)MAKE="$(MAKE)" NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" $(NODE_BIN) $(RUN_CHECK_SCHEDULE_SCRIPT) --target check --manifest "$(CHECK_SCHEDULE_MANIFEST)" --summary-profile check --resource-limit cpu=$(CHECK_JOBS) --resource-limit io=$(CHECK_IO_JOBS)

ci:
	$(Q)./scripts/ci/verify.sh

release-check: check run-harness-smoke-extended license-report sbom build

license-report:
	$(RUN_PHASE) "license-report" -- ./scripts/check-release-artifact.sh "license report" "$(LICENSE_REPORT_ARTIFACT)"

sbom:
	$(RUN_PHASE) "sbom" -- ./scripts/check-release-artifact.sh "SBOM" "$(SBOM_ARTIFACT)"

build: build-server build-migrate

$(EMBEDDED_WEB_ASSET_STAMP): $(CURDIR)/apps/web/dist/index.html
	$(Q)mkdir -p $(EMBEDDED_WEB_ASSET_DIR) $(dir $(EMBEDDED_WEB_ASSET_STAMP))
	$(Q)find $(EMBEDDED_WEB_ASSET_DIR) -mindepth 1 ! -name '.keep' -exec rm -rf {} +
	$(Q)cp -R $(CURDIR)/apps/web/dist/. $(EMBEDDED_WEB_ASSET_DIR)/
	$(Q)printf 'source=%s\n' "$(CURDIR)/apps/web/dist/index.html" > $(EMBEDDED_WEB_ASSET_STAMP)

$(SERVER_BIN): $$(SERVER_BUILD_INPUTS) $(EMBEDDED_WEB_ASSET_STAMP)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build server" -- $(GO_ENV) $(GO) build -o $(SERVER_BIN) ./cmd/server

build-server: $(SERVER_BIN)

$(MIGRATE_BIN): $$(MIGRATE_BUILD_INPUTS)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build migrate" -- $(GO_ENV) $(GO) build -o $(MIGRATE_BIN) ./cmd/migrate

build-migrate: $(MIGRATE_BIN)

$(CURDIR)/apps/web/dist/index.html: $$(WEB_BUILD_INPUTS) $(FRONTEND_INSTALL_STAMP) | $(NODE_BIN)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "build web" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vite build $(VITE_BUILD_FLAGS)

build-web: $(CURDIR)/apps/web/dist/index.html

clean:
	$(Q)$(call guarded_remove_paths,$(CLEAN_PATHS))
	$(Q)$(clean_embedded_web_assets)

distclean:
	$(Q)printf '%s\n' \
		'The following repo-local artifacts will be removed when present:' \
		'  $(SERVER_BIN)' \
		'  $(MIGRATE_BIN)' \
		'  $(CURDIR)/apps/web/dist' \
		'  $(EMBEDDED_WEB_ASSET_STAMP)' \
		'  $(CARTULARY_TEST_RESULTS_DIR)' \
		'  $(RELEASE_ARTIFACT_DIR)' \
		'  $(CURDIR)/apps/web/test-results' \
		'  $(CURDIR)/tmp/dev-stack' \
		'  $(CURDIR)/tmp/dev-stack-lifecycle-smoke.*' \
		'  $(CURDIR)/tmp/web-e2e-lifecycle-smoke.*' \
		'  $(CURDIR)/tmp/vitest-json-sample.*' \
		'  $(NODE_RUNTIME_DIR)' \
		'  $(TOOLBIN_DIR)' \
		'  $(CURDIR)/tmp/frontend-install' \
		'  $(CURDIR)/tmp/frontend-toolchain' \
		'  $(CURDIR)/tmp/playwright' \
		'  $(CURDIR)/tmp/frontend-embed' \
		'  $(CURDIR)/.cache' \
		'  $(CURDIR)/.pnpm-store' \
		'  $(CURDIR)/apps/web/.vite' \
		'  $(CURDIR)/playwright-report' \
		'  $(CURDIR)/apps/web/playwright-report' \
		'  $(CURDIR)/coverage' \
		'  $(CURDIR)/apps/web/coverage' \
		'  generated embedded web assets under $(EMBEDDED_WEB_ASSET_DIR), preserving .keep'
	$(Q)$(call guarded_remove_paths,$(DISTCLEAN_PATHS))
	$(Q)$(clean_embedded_web_assets)
