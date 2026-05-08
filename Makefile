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
HARNESS_SMOKE_JOBS ?= 6
BACKEND_STORE_GO_TEST_P ?= 2
BACKEND_INTEGRATION_GO_TEST_P ?= 2
BACKEND_INTEGRATION_SHARD_JOBS ?= 6
GO_TEST_SERVICE_PACKAGE_PARALLELISM ?= 1
PLAYWRIGHT_WORKERS ?= 3
BROWSER_E2E_FUNCTIONAL_SHARDS ?= auto
VITEST_MAX_WORKERS ?= 4
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
CYCLONEDX_GOMOD_BIN ?= $(TOOLBIN_DIR)/cyclonedx-gomod-v1.10.0
SYFT_BIN ?= $(TOOLBIN_DIR)/syft-v1.44.0
SHELLCHECK_VERSION ?= 0.11.0
SHELLCHECK_BIN ?= $(TOOLBIN_DIR)/shellcheck-v$(SHELLCHECK_VERSION)
SHELLCHECK_ARCHIVE_DIR ?= $(CURDIR)/tmp/shellcheck-archives
GOVULNCHECK_FLAGS ?= -test
GOVULNCHECK_PATTERNS ?= ./cmd/... ./internal/... ./db/... ./tools/...
GOVULNCHECK_DB ?=
GOSEC_RULES ?= G602,G124,G112,G114
GOSEC_FLAGS ?= -exclude-generated
GOSEC_PATTERNS ?= ./cmd/... ./internal/... ./db/... ./tools/...
GOSEC_TARGETED_RUNTIME_RULES ?= G122,G301,G302,G303,G304,G305,G306,G307
GOSEC_TARGETED_RUNTIME_FLAGS ?= -exclude-generated -quiet -exclude-dir=internal/testutil
GOSEC_TARGETED_RUNTIME_PATTERNS ?= ./cmd/... ./internal/...
GOSEC_AUDIT_RUNTIME_RULES ?= G118,G122,G301,G302,G303,G304,G305,G306,G307
GOSEC_AUDIT_RUNTIME_FLAGS ?= -exclude-generated -no-fail -quiet -exclude-dir=internal/testutil
GOSEC_AUDIT_RUNTIME_PATTERNS ?= ./cmd/... ./internal/...
GOSEC_AUDIT_SUPPORT_RULES ?= G122,G301,G302,G303,G304,G305,G306,G307
GOSEC_AUDIT_SUPPORT_FLAGS ?= -exclude-generated -no-fail -quiet
GOSEC_AUDIT_SUPPORT_PATTERNS ?= ./internal/testutil/... ./tools/...
TEST_SERVICES_BIN ?= $(TOOLBIN_DIR)/cartulary-test-services
SERVICE_BACKED_SCHEDULE_MANIFEST ?= $(CURDIR)/tools/service_backed_schedule_manifest.json
EXECUTION_TOPOLOGY_MANIFEST ?= $(CURDIR)/tools/execution_topology_manifest.json
GO_TEST_DURATION_BASELINE ?= $(CURDIR)/tools/go_test_duration_baselines.json
BROWSER_E2E_DURATION_BASELINE ?= $(CURDIR)/tools/browser_e2e_duration_baselines.json
SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE ?= $(CURDIR)/tools/service_backed_make_target_duration_baselines.json
HARNESS_SMOKE_DURATION_BASELINE ?= $(CURDIR)/tools/harness_smoke_duration_baselines.json
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
TEST_OUTPUT_SCRIPT := $(CURDIR)/scripts/lib/test-output.mjs
CARTULARY_RUNNER_SCRIPT := $(CURDIR)/scripts/cartulary-runner.mjs
TASK_SURFACE_MANIFEST ?= $(CURDIR)/tools/task_surface_manifest.json
TASK_SURFACE_REPORT_ARGS ?=
RUN_MAKE_SEQUENCE_SCRIPT := $(CURDIR)/scripts/run-make-sequence.sh
RUN_HARNESS_SMOKE_SCRIPT := $(CURDIR)/scripts/run-harness-smoke.mjs
RUN_SERVICE_BACKED_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-service-backed-schedule.mjs
RUN_CHECK_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-check-schedule.mjs
BUILD_INPUTS_SCRIPT := $(CURDIR)/scripts/list-build-inputs.sh
RUN_PHASE = $(Q)$(RUN_PHASE_SCRIPT)

define resolve_service_go_test_p
$(if $(filter environment environment override command line override,$(origin $(1))),$($(1)),$(if $(filter environment environment override command line override,$(origin GO_TEST_SERVICE_PACKAGE_PARALLELISM)),$(GO_TEST_SERVICE_PACKAGE_PARALLELISM),$($(1))))
endef

EFFECTIVE_BACKEND_STORE_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_STORE_GO_TEST_P)
EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_INTEGRATION_GO_TEST_P)

CARTULARY_OUTPUT_MODE ?= summary
CARTULARY_TEST_RESULTS_DIR ?= $(CURDIR)/.cartulary/test-results
CARTULARY_TEST_RUN_ID ?= $(shell if [ -x /usr/bin/date ]; then now="$$(/usr/bin/date -u +%Y%m%dT%H%M%SZ)"; elif command -v date >/dev/null 2>&1; then now="$$(date -u +%Y%m%dT%H%M%SZ)"; else now="unknown-time"; fi; printf '%s-p%s' "$$now" "$$$$")
RELEASE_ARTIFACT_DIR ?= $(CURDIR)/.cartulary/release-artifacts
LICENSE_REPORT_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/license-report.json
SBOM_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/sbom.cyclonedx.json
BENCHMARK_MANIFEST ?= $(CURDIR)/.cartulary/benchmark/benchmark_manifest.json
export CARTULARY_OUTPUT_MODE VERBOSE CI_VERBOSE CARTULARY_TEST_RESULTS_DIR CARTULARY_TEST_RUN_ID CARTULARY_TEST_INVENTORY

ifeq ($(CI_VERBOSE),1)
EFFECTIVE_OUTPUT_MODE := normal
else ifeq ($(VERBOSE),1)
EFFECTIVE_OUTPUT_MODE := normal
else
EFFECTIVE_OUTPUT_MODE := $(CARTULARY_OUTPUT_MODE)
endif

ifneq ($(filter $(EFFECTIVE_OUTPUT_MODE),quiet summary ci machine),)
PNPM_INSTALL_FLAGS := --reporter=append-only --loglevel=warn
BIOME_CHECK_FLAGS := --reporter=summary --diagnostic-level=warn
TSC_FLAGS := --pretty false
VITE_BUILD_FLAGS := --logLevel warn
VITEST_FLAGS := --silent=passed-only
VITEST_MANIFEST_FLAGS := --silent=passed-only
PLAYWRIGHT_TEST_FLAGS := --quiet
export NO_COLOR := 1
export CLICOLOR := 0
export FORCE_COLOR := 0
endif

SQLC_TOOL := github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
GOOSE_TOOL := github.com/pressly/goose/v3/cmd/goose@v3.27.0
STATICCHECK_TOOL := honnef.co/go/tools/cmd/staticcheck@v0.7.0
GOVULNCHECK_TOOL := golang.org/x/vuln/cmd/govulncheck@v1.3.0
GOSEC_TOOL := github.com/securego/gosec/v2/cmd/gosec@v2.26.1
CYCLONEDX_GOMOD_TOOL := github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.10.0
SYFT_TOOL := github.com/anchore/syft/cmd/syft@v1.44.0
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
CLEAN_PATHS := $(SERVER_BIN) $(MIGRATE_BIN) $(CURDIR)/apps/web/dist $(EMBEDDED_WEB_ASSET_STAMP) $(CARTULARY_TEST_RESULTS_DIR) $(RELEASE_ARTIFACT_DIR) $(CURDIR)/test-results $(CURDIR)/apps/web/test-results $(CURDIR)/playwright-report $(CURDIR)/apps/web/playwright-report $(CURDIR)/coverage $(CURDIR)/apps/web/coverage $(CURDIR)/.vite $(CURDIR)/apps/web/.vite $(CURDIR)/node_modules/.vite* $(CURDIR)/apps/web/node_modules/.vite* $(CURDIR)/packages/*/node_modules/.vite*
DISTCLEAN_PATHS := $(CLEAN_PATHS) $(NODE_RUNTIME_DIR) $(CURDIR)/tmp/node-archives $(TOOLBIN_DIR) $(SHELLCHECK_ARCHIVE_DIR) $(CURDIR)/tmp/frontend-install $(CURDIR)/tmp/frontend-toolchain $(CURDIR)/tmp/playwright $(CURDIR)/tmp/frontend-embed $(CURDIR)/.cache $(CURDIR)/.pnpm-store
CLEAN_TMP_PRESERVE_NAMES := node-runtime node-archives toolbin shellcheck-archives frontend-install frontend-toolchain playwright frontend-embed

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

define clean_tmp_scratch
set -euo pipefail; \
repo="$(CURDIR)"; \
dir="$(CURDIR)/tmp"; \
if [ ! -d "$$dir" ]; then \
	exit 0; \
fi; \
case "$$dir" in \
	"$$repo"/*) ;; \
	*) echo "refusing tmp cleanup path outside repository: $$dir" >&2; exit 1 ;; \
esac; \
for path in "$$dir"/* "$$dir"/.[!.]* "$$dir"/..?*; do \
	if [ ! -e "$$path" ] && [ ! -L "$$path" ]; then \
		continue; \
	fi; \
	name="$${path##*/}"; \
	case " $(CLEAN_TMP_PRESERVE_NAMES) " in \
		*" $$name "*) continue ;; \
	esac; \
	printf 'removing %s\n' "$${path#$$repo/}"; \
	rm -rf -- "$$path"; \
done
endef

define print_cleanup_paths
set -euo pipefail; \
printf '%s\n' 'The following repo-local artifacts will be removed when present:'; \
for path in $(1); do \
	printf '  %s\n' "$$path"; \
done; \
printf '  %s\n' 'repo-local scratch directories under $(CURDIR)/tmp'; \
printf '  %s\n' 'generated embedded web assets under $(EMBEDDED_WEB_ASSET_DIR), preserving .keep'
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

$(SBOM_ARTIFACT) $(LICENSE_REPORT_ARTIFACT): $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) $(CYCLONEDX_GOMOD_BIN) $(SYFT_BIN) scripts/generate-sbom-license-evidence.mjs scripts/validate-cyclonedx.mjs go.mod go.sum package.json pnpm-lock.yaml pnpm-workspace.yaml docker-compose.dev.yml $(wildcard apps/web/package.json packages/*/package.json)
	$(Q)mkdir -p $(RELEASE_ARTIFACT_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)CARTULARY_TEST_TARGET=release-evidence CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_PHASE_SCRIPT) "generate SBOM/license evidence" -- env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack" GO="$(GO)" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" NODE_BIN="$(NODE_BIN)" PNPM="$(PNPM)" CYCLONEDX_GOMOD_BIN="$(CYCLONEDX_GOMOD_BIN)" SYFT_BIN="$(SYFT_BIN)" RELEASE_ARTIFACT_DIR="$(RELEASE_ARTIFACT_DIR)" LICENSE_REPORT_ARTIFACT="$(LICENSE_REPORT_ARTIFACT)" SBOM_ARTIFACT="$(SBOM_ARTIFACT)" $(NODE_BIN) ./scripts/generate-sbom-license-evidence.mjs

$(NODE_BIN): scripts/bootstrap-node-runtime.sh Makefile
	$(Q)NODE_VERSION="$(NODE_VERSION)" NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" ./scripts/bootstrap-node-runtime.sh

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

$(FRONTEND_INSTALL_STAMP): $(FRONTEND_INSTALL_INPUTS) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(FRONTEND_INSTALL_STAMP))
	$(Q)CARTULARY_TEST_TARGET=frontend-install CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_PHASE_SCRIPT) "frontend install" -- $(PNPM_ENV) $(PNPM) install $(PNPM_INSTALL_FLAGS)
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

$(CYCLONEDX_GOMOD_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/cyclonedx-gomod $(CYCLONEDX_GOMOD_BIN)
	$(RUN_PHASE) "bootstrap cyclonedx-gomod tool" -- bash -c 'cd "$$1" && env GOBIN="$$2" GOCACHE="$$3" GOMODCACHE="$$4" "$$5" install "$$6"' _ "$(GO_CACHE_DIR)" "$(TOOLBIN_DIR)" "$(GO_CACHE_DIR)" "$(GO_MOD_CACHE_DIR)" "$(GO)" "$(CYCLONEDX_GOMOD_TOOL)"
	$(Q)mv $(TOOLBIN_DIR)/cyclonedx-gomod $(CYCLONEDX_GOMOD_BIN)

$(SYFT_BIN):
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)rm -f $(TOOLBIN_DIR)/syft $(SYFT_BIN)
	$(RUN_PHASE) "bootstrap syft tool" -- bash -c 'cd "$$1" && env GOBIN="$$2" GOCACHE="$$3" GOMODCACHE="$$4" "$$5" install "$$6"' _ "$(GO_CACHE_DIR)" "$(TOOLBIN_DIR)" "$(GO_CACHE_DIR)" "$(GO_MOD_CACHE_DIR)" "$(GO)" "$(SYFT_TOOL)"
	$(Q)mv $(TOOLBIN_DIR)/syft $(SYFT_BIN)

$(SHELLCHECK_BIN): scripts/bootstrap-shellcheck.sh Makefile
	$(RUN_PHASE) "bootstrap shellcheck tool" -- env SHELLCHECK_VERSION=$(SHELLCHECK_VERSION) TOOLBIN_DIR=$(TOOLBIN_DIR) SHELLCHECK_BIN=$(SHELLCHECK_BIN) CARTULARY_SHELLCHECK_ARCHIVE_DIR=$(SHELLCHECK_ARCHIVE_DIR) ./scripts/bootstrap-shellcheck.sh

$(TEST_SERVICES_BIN): $$(TEST_SERVICES_BUILD_INPUTS)
	$(Q)mkdir -p $(TOOLBIN_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)CARTULARY_TEST_TARGET=testservices-build CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_PHASE_SCRIPT) "build testservices" -- $(GO_ENV) $(GO) build -o $(TEST_SERVICES_BIN) ./tools/testservices

$(PLAYWRIGHT_INSTALL_STAMP): $(FRONTEND_INSTALL_STAMP) $(FRONTEND_TOOLCHAIN_STAMP)
	$(Q)mkdir -p $(dir $(PLAYWRIGHT_INSTALL_STAMP))
	$(RUN_PHASE) "playwright-install" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec playwright install chromium
	$(Q)printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' "$(NODE_BIN)" "$(NODE_VERSION)" "$(PNPM)" "$(PNPM_VERSION)" > $(PLAYWRIGHT_INSTALL_STAMP)

$(EMBEDDED_WEB_ASSET_STAMP): $(CURDIR)/apps/web/dist/index.html
	$(Q)mkdir -p $(EMBEDDED_WEB_ASSET_DIR) $(dir $(EMBEDDED_WEB_ASSET_STAMP))
	$(Q)find $(EMBEDDED_WEB_ASSET_DIR) -mindepth 1 ! -name '.keep' -exec rm -rf {} +
	$(Q)cp -R $(CURDIR)/apps/web/dist/. $(EMBEDDED_WEB_ASSET_DIR)/
	$(Q)printf 'source=%s\n' "$(CURDIR)/apps/web/dist/index.html" > $(EMBEDDED_WEB_ASSET_STAMP)

$(SERVER_BIN): $$(SERVER_BUILD_INPUTS) $(EMBEDDED_WEB_ASSET_STAMP)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build server" -- $(GO_ENV) $(GO) build -o $(SERVER_BIN) ./cmd/server

$(MIGRATE_BIN): $$(MIGRATE_BUILD_INPUTS)
	$(Q)mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(RUN_PHASE) "build migrate" -- $(GO_ENV) $(GO) build -o $(MIGRATE_BIN) ./cmd/migrate

$(CURDIR)/apps/web/dist/index.html: $$(WEB_BUILD_INPUTS) $(FRONTEND_INSTALL_STAMP) | $(NODE_BIN)
	$(Q)CARTULARY_TEST_TARGET=build-web CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_PHASE_SCRIPT) "build web" -- $(PNPM_ENV) $(PNPM) --dir apps/web exec vite build $(VITE_BUILD_FLAGS)
