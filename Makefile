SHELL := /bin/bash
.DEFAULT_GOAL := help

.SECONDEXPANSION:

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
GO_CACHE_DIR ?= /tmp/cartulary-go-build
GO_MOD_CACHE_DIR ?= /tmp/cartulary-go-mod
GO_RUN_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
CHECK_HOST_CPU_JOBS ?= 8
CHECK_HOST_IO_JOBS ?= 12
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
STATICCHECK_BIN ?= $(TOOLBIN_DIR)/staticcheck-v0.7.0
GOVULNCHECK_BIN ?= $(TOOLBIN_DIR)/govulncheck-v1.3.0
GOSEC_BIN ?= $(TOOLBIN_DIR)/gosec-v2.26.1
SHELLCHECK_VERSION ?= 0.11.0
SHELLCHECK_BIN ?= $(TOOLBIN_DIR)/shellcheck-v$(SHELLCHECK_VERSION)
SHELLCHECK_ARCHIVE_DIR ?= $(CURDIR)/tmp/shellcheck-archives
GOVULNCHECK_FLAGS ?= -test
GOVULNCHECK_PATTERNS ?= ./...
GOVULNCHECK_DB ?=
GOSEC_RULES ?= G602,G124,G112,G114
GOSEC_FLAGS ?= -exclude-generated
GOSEC_PATTERNS ?= ./cmd/... ./internal/... ./db/... ./tools/...
GOSEC_AUDIT_RUNTIME_RULES ?= G118,G304,G122,G301,G302,G306,G307
GOSEC_AUDIT_RUNTIME_FLAGS ?= -exclude-generated -no-fail -quiet -exclude-dir=internal/testutil
GOSEC_AUDIT_RUNTIME_PATTERNS ?= ./cmd/... ./internal/...
GOSEC_AUDIT_SUPPORT_RULES ?= G304,G122,G301,G302,G306,G307
GOSEC_AUDIT_SUPPORT_FLAGS ?= -exclude-generated -no-fail -quiet
GOSEC_AUDIT_SUPPORT_PATTERNS ?= ./internal/testutil/... ./tools/...
TEST_SERVICES_BIN ?= $(TOOLBIN_DIR)/cartulary-test-services
SERVICE_BACKED_SCHEDULE_MANIFEST ?= $(CURDIR)/tools/service_backed_schedule_manifest.json
EXECUTION_TOPOLOGY_MANIFEST ?= $(CURDIR)/tools/execution_topology_manifest.json
SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE ?= $(CURDIR)/tools/service_backed_make_target_duration_baselines.json
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
RUN_SCRIPTS_BIOME_SCRIPT := $(CURDIR)/scripts/run-scripts-biome.sh
TEST_OUTPUT_SCRIPT := $(CURDIR)/scripts/lib/test-output.mjs
CARTULARY_RUNNER_SCRIPT := $(CURDIR)/scripts/cartulary-runner.mjs
TASK_SURFACE_MANIFEST ?= $(CURDIR)/tools/task_surface_manifest.json
TASK_SURFACE_REPORT_ARGS ?=
RUN_MAKE_SEQUENCE_SCRIPT := $(CURDIR)/scripts/run-make-sequence.sh
RUN_HARNESS_SMOKE_SCRIPT := $(CURDIR)/scripts/run-harness-smoke.mjs
RUN_SERVICE_BACKED_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-service-backed-schedule.mjs
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
TARGET_SUMMARY = $(Q)NODE_BIN=$(NODE_BIN) TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) target-summary

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
STATICCHECK_TOOL := honnef.co/go/tools/cmd/staticcheck@v0.7.0
GOVULNCHECK_TOOL := golang.org/x/vuln/cmd/govulncheck@v1.3.0
GOSEC_TOOL := github.com/securego/gosec/v2/cmd/gosec@v2.26.1
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
DISTCLEAN_PATHS := $(CLEAN_PATHS) $(NODE_RUNTIME_DIR) $(TOOLBIN_DIR) $(SHELLCHECK_ARCHIVE_DIR) $(CURDIR)/tmp/frontend-install $(CURDIR)/tmp/frontend-toolchain $(CURDIR)/tmp/playwright $(CURDIR)/tmp/frontend-embed $(CURDIR)/.cache $(CURDIR)/.pnpm-store $(CURDIR)/apps/web/.vite $(CURDIR)/playwright-report $(CURDIR)/apps/web/playwright-report $(CURDIR)/coverage $(CURDIR)/apps/web/coverage

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

include tools/task_surface.generated.mk

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
	shellcheck_path=""; \
	if [ -x "$(SHELLCHECK_BIN)" ]; then \
		shellcheck_path="$(SHELLCHECK_BIN)"; \
	fi; \
	if [ -n "$$shellcheck_path" ]; then \
		shellcheck_version="$$("$$shellcheck_path" --version 2>/dev/null | awk -F': ' '$$1 == "version" { print $$2; exit }')"; \
		if [ "$$shellcheck_version" = "$(SHELLCHECK_VERSION)" ]; then \
			printf 'ok shellcheck: %s %s\n' "$$shellcheck_path" "$$shellcheck_version"; \
		else \
			printf 'missing shellcheck: expected %s, found %s at %s\n' "$(SHELLCHECK_VERSION)" "$${shellcheck_version:-unusable}" "$$shellcheck_path"; \
			fail=1; \
		fi; \
	else \
		print_missing shellcheck "run make shell-lint-toolchain"; \
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

$(STATICCHECK_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/staticcheck $(STATICCHECK_BIN)
	$(RUN_PHASE) "bootstrap staticcheck tool" -- env GOBIN=$(TOOLBIN_DIR) $(GO_RUN_ENV) $(GO) install $(STATICCHECK_TOOL)
	$(Q)mv $(TOOLBIN_DIR)/staticcheck $(STATICCHECK_BIN)

$(GOVULNCHECK_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/govulncheck $(GOVULNCHECK_BIN)
	$(RUN_PHASE) "bootstrap govulncheck tool" -- env GOBIN=$(TOOLBIN_DIR) $(GO_RUN_ENV) $(GO) install $(GOVULNCHECK_TOOL)
	$(Q)mv $(TOOLBIN_DIR)/govulncheck $(GOVULNCHECK_BIN)

$(GOSEC_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/gosec $(GOSEC_BIN)
	$(RUN_PHASE) "bootstrap gosec tool" -- env GOBIN=$(TOOLBIN_DIR) $(GO_RUN_ENV) $(GO) install $(GOSEC_TOOL)
	$(Q)mv $(TOOLBIN_DIR)/gosec $(GOSEC_BIN)

$(SHELLCHECK_BIN): scripts/bootstrap-shellcheck.sh Makefile
	$(RUN_PHASE) "bootstrap shellcheck tool" -- env SHELLCHECK_VERSION=$(SHELLCHECK_VERSION) TOOLBIN_DIR=$(TOOLBIN_DIR) SHELLCHECK_BIN=$(SHELLCHECK_BIN) CARTULARY_SHELLCHECK_ARCHIVE_DIR=$(SHELLCHECK_ARCHIVE_DIR) ./scripts/bootstrap-shellcheck.sh

$(TEST_SERVICES_BIN): $$(TEST_SERVICES_BUILD_INPUTS)
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build testservices" -- $(GO_ENV) $(GO) build -o $(TEST_SERVICES_BIN) ./tools/testservices

$(PLAYWRIGHT_INSTALL_STAMP): $(FRONTEND_INSTALL_STAMP) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(PLAYWRIGHT_INSTALL_STAMP))
	$(RUN_PHASE) "playwright-install" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec playwright install chromium
	$(Q)printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' "$(NODE_BIN)" "$(NODE_VERSION)" "$(PNPM)" "$(PNPM_VERSION)" > $(PLAYWRIGHT_INSTALL_STAMP)

bootstrap: $(SQLC_BIN) $(GOOSE_BIN) $(STATICCHECK_BIN) $(GOVULNCHECK_BIN) $(GOSEC_BIN) $(SHELLCHECK_BIN) frontend-install playwright-install
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

go-lint-toolchain: $(STATICCHECK_BIN)

go-security-toolchain: $(GOVULNCHECK_BIN) $(GOSEC_BIN)

shell-lint-toolchain: $(SHELLCHECK_BIN)

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
	$(RUN_PHASE) "phase-schedules" -- $(NODE_BIN) ./scripts/render-execution-topology-artifacts.mjs --topology "$(EXECUTION_TOPOLOGY_MANIFEST)"

phase-schedule-drift: $(NODE_BIN)
	$(RUN_PHASE) "phase-schedule-drift" -- $(NODE_BIN) ./scripts/render-execution-topology-artifacts.mjs --check --topology "$(EXECUTION_TOPOLOGY_MANIFEST)"

benchmark-claim-check: $(NODE_BIN)
	$(RUN_PHASE) "benchmark-claim-check" -- $(NODE_BIN) ./scripts/check-benchmark-claim.mjs "$(BENCHMARK_MANIFEST)"

test-service-images: $(TEST_SERVICES_BIN)
	$(RUN_PHASE) "warm test service images" -- $(TEST_SERVICES_BIN) warm-images

test-local: backend-unit frontend-typecheck frontend-unit

frontend-unit: export CARTULARY_TEST_TARGET := frontend-unit

frontend-typecheck: export CARTULARY_TEST_TARGET := frontend-typecheck

frontend-typecheck: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "frontend typecheck" -- $(PNPM_ENV) $(PNPM) typecheck
	$(TARGET_SUMMARY) frontend-typecheck pass

frontend-unit: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)env PNPM=$(PNPM) NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) VITEST_FLAGS="$(VITEST_FLAGS)" VITEST_MAX_WORKERS=$(VITEST_MAX_WORKERS) ./scripts/run-frontend-unit.sh

lint: lint-go lint-biome frontend-import-boundary-check lint-scripts lint-shell frontend-typecheck

lint-go: lint-go-format lint-go-vet lint-go-staticcheck

lint-go-format:
	$(Q)CARTULARY_PHASE_FAILURE_NOTE="run make format to apply authored Go formatting" $(RUN_PHASE_SCRIPT) "lint gofmt" -- bash ./scripts/run-go-format.sh --check

lint-go-vet:
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "lint go-vet" -- bash ./scripts/run-go-vet.sh

lint-go-staticcheck: go-lint-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "lint staticcheck" -- env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) STATICCHECK_BIN=$(STATICCHECK_BIN) STATICCHECK_CHECKS="$(STATICCHECK_CHECKS)" bash ./scripts/run-go-staticcheck.sh

go-vulncheck: go-security-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "go vulncheck" -- env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) GOVULNCHECK_BIN=$(GOVULNCHECK_BIN) GOVULNCHECK_FLAGS="$(GOVULNCHECK_FLAGS)" GOVULNCHECK_PATTERNS="$(GOVULNCHECK_PATTERNS)" GOVULNCHECK_DB="$(GOVULNCHECK_DB)" bash ./scripts/run-go-govulncheck.sh

go-gosec-targeted: go-security-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "go gosec targeted" -- env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) GOSEC_BIN=$(GOSEC_BIN) GOSEC_RULES="$(GOSEC_RULES)" GOSEC_FLAGS="$(GOSEC_FLAGS)" GOSEC_PATTERNS="$(GOSEC_PATTERNS)" bash ./scripts/run-go-gosec-targeted.sh

go-gosec-audit: go-security-toolchain
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "go gosec audit" -- env GO=$(GO) GO_CACHE_DIR=$(GO_CACHE_DIR) GO_MOD_CACHE_DIR=$(GO_MOD_CACHE_DIR) GOSEC_BIN=$(GOSEC_BIN) GOSEC_AUDIT_RUNTIME_RULES="$(GOSEC_AUDIT_RUNTIME_RULES)" GOSEC_AUDIT_RUNTIME_FLAGS="$(GOSEC_AUDIT_RUNTIME_FLAGS)" GOSEC_AUDIT_RUNTIME_PATTERNS="$(GOSEC_AUDIT_RUNTIME_PATTERNS)" GOSEC_AUDIT_SUPPORT_RULES="$(GOSEC_AUDIT_SUPPORT_RULES)" GOSEC_AUDIT_SUPPORT_FLAGS="$(GOSEC_AUDIT_SUPPORT_FLAGS)" GOSEC_AUDIT_SUPPORT_PATTERNS="$(GOSEC_AUDIT_SUPPORT_PATTERNS)" bash ./scripts/run-go-gosec-audit.sh

format: format-go format-frontend

format-go:
	$(Q)$(RUN_PHASE_SCRIPT) "format go" -- bash ./scripts/run-go-format.sh --write

format-frontend: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT) "format frontend" -- bash $(RUN_FRONTEND_BIOME_SCRIPT) format

lint-biome: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)CARTULARY_PHASE_FAILURE_NOTE="inspect Biome diagnostics; run make format only for formatting/style diagnostics" CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT) "lint biome" -- bash $(RUN_FRONTEND_BIOME_SCRIPT) check $(BIOME_CHECK_FLAGS)

lint-scripts: $(NODE_BIN) $(FRONTEND_INSTALL_STAMP)
	$(Q)CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 $(RUN_PHASE_SCRIPT) "lint scripts" -- bash $(RUN_SCRIPTS_BIOME_SCRIPT) $(BIOME_SCRIPT_CHECK_FLAGS)

lint-shell: shell-lint-toolchain
	$(RUN_PHASE_ALLOW_SUCCESS_LOG) "lint shell" -- env SHELLCHECK_BIN=$(SHELLCHECK_BIN) LINT_SHELL_STRICT="$(LINT_SHELL_STRICT)" ./scripts/run-shellcheck.sh

check-setup-blockers: $(NODE_BIN)
	$(Q)$(MAKE) --no-print-directory toolchain-drift
	$(Q)$(MAKE) --no-print-directory codegen-toolchain
	$(Q)$(MAKE) --no-print-directory go-lint-toolchain
	$(Q)$(MAKE) --no-print-directory shell-lint-toolchain
	$(Q)$(MAKE) --no-print-directory go-security-toolchain
	$(Q)if [ "$(CI)" = "1" ]; then \
		$(MAKE) --no-print-directory frontend-install-ci; \
	else \
		$(MAKE) --no-print-directory frontend-install; \
	fi

check-harness-smoke:
	$(Q)NODE_BIN=$(NODE_BIN) TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) summary-target --target check-harness-smoke --child-target run-harness-smoke-fast --status pass --phase-label check-harness-smoke --projection check-harness-smoke

# Build and service-image readiness gate the service-backed scheduler. Keep
# independent static and harness validation outside this prerequisite aggregate
# so service-backed work can start as soon as these artifacts are ready.
check-build-prereqs: build-server build-migrate test-service-images

check-frontend-unit: frontend-unit

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
		'  $(SHELLCHECK_ARCHIVE_DIR)' \
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
