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
OPERATOR_BIN ?= $(CURDIR)/operator
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
SCHEDULER_MANIFEST ?= $(CURDIR)/tools/scheduler_manifest.json
EXECUTION_TOPOLOGY_MANIFEST ?= $(CURDIR)/tools/execution_topology_manifest.json
GO_TEST_DURATION_BASELINE ?= $(CURDIR)/tools/go_test_duration_baselines.json
BROWSER_E2E_DURATION_BASELINE ?= $(CURDIR)/tools/browser_e2e_duration_baselines.json
SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE ?= $(CURDIR)/tools/service_backed_make_target_duration_baselines.json
HARNESS_SMOKE_DURATION_BASELINE ?= $(CURDIR)/tools/harness_smoke_duration_baselines.json
OBJECT_STORE_BUCKET ?= cartulary
FRONTEND_INSTALL_STAMP ?= $(CURDIR)/tmp/frontend-install/node-v$(NODE_VERSION)-pnpm-v$(PNPM_VERSION).stamp
PLAYWRIGHT_INSTALL_STAMP ?= $(CURDIR)/tmp/playwright/chromium.stamp
FRONTEND_TOOLCHAIN_STAMP ?= $(CURDIR)/tmp/frontend-toolchain/node-v$(NODE_VERSION)-pnpm-v$(PNPM_VERSION).stamp
CARTULARY_READINESS_CACHE_DIR ?= $(CURDIR)/.cache/cartulary/readiness
CARTULARY_BUILD_CACHE_DIR ?= $(CURDIR)/.cache/cartulary/build-artifacts
FRONTEND_NODE_MODULES_DIRS ?= $(CURDIR)/node_modules $(CURDIR)/apps/web/node_modules $(CURDIR)/packages/*/node_modules
PNPM_RUN_ENV := PATH=$(NODE_RUNTIME_DIR)/bin:$$PATH COREPACK_HOME=$(NODE_RUNTIME_DIR)/corepack
GO_ENV := env $(GO_RUN_ENV)
PNPM_ENV := env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack"
BROWSER_E2E_OWNED_STACK_ENV := PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" NODE_RUNTIME_DIR=$(NODE_RUNTIME_DIR) NODE_BIN=$(NODE_BIN) PNPM=$(PNPM) CARTULARY_SERVER_BIN=$(SERVER_BIN) CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN) CARTULARY_TEST_SERVICES_BIN=$(TEST_SERVICES_BIN) CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES=1
Q := @
comma := ,
RUN_PHASE_SCRIPT := $(CURDIR)/tools/harness/core/run-phase.sh
TEST_OUTPUT_SCRIPT := $(CURDIR)/tools/harness/core/test-output.mjs
CARTULARY_RUNNER_SCRIPT := $(CURDIR)/scripts/cartulary-runner.mjs
TASK_SURFACE_MANIFEST ?= $(CURDIR)/tools/task_surface_manifest.json
TASK_SURFACE_REPORT_ARGS ?=
RUN_MAKE_SEQUENCE_SCRIPT := $(CURDIR)/scripts/run-make-sequence.sh
RUN_HARNESS_SMOKE_SCRIPT := $(CURDIR)/scripts/run-harness-smoke.mjs
RUN_SERVICE_BACKED_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-service-backed-schedule.mjs
RUN_CHECK_SCHEDULE_SCRIPT := $(CURDIR)/scripts/run-check-schedule.mjs
BUILD_INPUTS_SCRIPT := $(CURDIR)/scripts/list-build-inputs.sh
CACHE_ARTIFACT_SCRIPT := $(CURDIR)/scripts/cache-artifact.sh
HARNESS_CONTRACT_SCRIPT := $(CURDIR)/scripts/harness-contract.sh
RUN_PHASE = $(Q)$(RUN_PHASE_SCRIPT)
RUN_HARNESS_PREFLIGHT = $(Q)$(HARNESS_CONTRACT_SCRIPT) preflight
RUN_HARNESS_CLEANUP = $(Q)$(HARNESS_CONTRACT_SCRIPT) cleanup

define resolve_service_go_test_p
$(if $(filter environment environment override command line override,$(origin $(1))),$($(1)),$(if $(filter environment environment override command line override,$(origin GO_TEST_SERVICE_PACKAGE_PARALLELISM)),$(GO_TEST_SERVICE_PACKAGE_PARALLELISM),$($(1))))
endef

EFFECTIVE_BACKEND_STORE_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_STORE_GO_TEST_P)
EFFECTIVE_BACKEND_INTEGRATION_GO_TEST_P := $(call resolve_service_go_test_p,BACKEND_INTEGRATION_GO_TEST_P)

DEFAULT_CARTULARY_TEST_RESULTS_DIR := $(CURDIR)/.cartulary/test-results
CARTULARY_OUTPUT_MODE ?=
CARTULARY_TEST_RESULTS_DIR ?= $(DEFAULT_CARTULARY_TEST_RESULTS_DIR)
CARTULARY_TEST_RUN_ID ?= $(shell if [ -x /usr/bin/date ]; then now="$$(/usr/bin/date -u +%Y%m%dT%H%M%SZ)"; elif command -v date >/dev/null 2>&1; then now="$$(date -u +%Y%m%dT%H%M%SZ)"; else now="unknown-time"; fi; printf '%s-p%s' "$$now" "$$$$")
RELEASE_ARTIFACT_DIR ?= $(CURDIR)/.cartulary/release-artifacts
LICENSE_REPORT_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/license-report.json
SBOM_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/sbom.cyclonedx.json
BENCHMARK_MANIFEST ?= $(CURDIR)/.cartulary/benchmark/benchmark_manifest.json
export CARTULARY_OUTPUT_MODE VERBOSE CI_VERBOSE CI CARTULARY_TEST_RESULTS_DIR CARTULARY_TEST_RUN_ID CARTULARY_TEST_INVENTORY

ifneq ($(strip $(CARTULARY_OUTPUT_MODE)),)
EFFECTIVE_OUTPUT_MODE := $(strip $(CARTULARY_OUTPUT_MODE))
else ifeq ($(VERBOSE),1)
EFFECTIVE_OUTPUT_MODE := verbose
else ifeq ($(CI_VERBOSE),1)
EFFECTIVE_OUTPUT_MODE := ci
else ifeq ($(CI),1)
EFFECTIVE_OUTPUT_MODE := ci
else
EFFECTIVE_OUTPUT_MODE := summary
endif

ifneq ($(filter $(EFFECTIVE_OUTPUT_MODE),quiet summary ci machine),)
PNPM_INSTALL_FLAGS := --reporter=append-only --loglevel=warn
BIOME_CHECK_FLAGS := --reporter=summary --diagnostic-level=warn
TSC_FLAGS := --pretty false
VITE_BUILD_FLAGS := --logLevel warn --sourcemap
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
OPERATOR_BUILD_INPUTS = go.mod go.sum $(call discover_build_inputs,cmd/operator internal/app internal/modules internal/platform contracts db/migrations tools/migration_history_manifest.json)
WEB_BUILD_INPUTS = package.json pnpm-lock.yaml pnpm-workspace.yaml $(call discover_build_inputs,apps/web packages)
TEST_SERVICES_BUILD_INPUTS = go.mod go.sum $(call discover_build_inputs,tools/testservices internal/testutil/pgtest internal/testutil/s3test internal/testutil/suiteservices internal/platform/postgres db/migrations)
WEB_DIST_INDEX := $(CURDIR)/apps/web/dist/index.html
EMBEDDED_WEB_ASSET_DIR := $(CURDIR)/internal/platform/httpapi/webassets/dist
EMBEDDED_WEB_ASSET_ARCHIVE := $(EMBEDDED_WEB_ASSET_DIR)/web-assets.zip
EMBEDDED_WEB_ASSET_READY_STAMP := $(CURDIR)/tmp/frontend-embed/web-assets.ready
EMBEDDED_WEB_ASSET_STAMP := $(CURDIR)/tmp/frontend-embed/web-assets.stamp
CLEAN_PATHS := $(SERVER_BIN) $(MIGRATE_BIN) $(OPERATOR_BIN) $(CURDIR)/apps/web/dist $(EMBEDDED_WEB_ASSET_STAMP) $(EMBEDDED_WEB_ASSET_READY_STAMP) $(DEFAULT_CARTULARY_TEST_RESULTS_DIR) $(RELEASE_ARTIFACT_DIR) $(CURDIR)/test-results $(CURDIR)/apps/web/test-results $(CURDIR)/playwright-report $(CURDIR)/apps/web/playwright-report $(CURDIR)/coverage $(CURDIR)/apps/web/coverage $(CURDIR)/.vite $(CURDIR)/apps/web/.vite $(CURDIR)/node_modules/.vite* $(CURDIR)/apps/web/node_modules/.vite* $(CURDIR)/packages/*/node_modules/.vite*
DISTCLEAN_PATHS := $(CLEAN_PATHS) $(NODE_RUNTIME_DIR) $(CURDIR)/tmp/node-archives $(TOOLBIN_DIR) $(SHELLCHECK_ARCHIVE_DIR) $(CURDIR)/tmp/frontend-install $(CURDIR)/tmp/frontend-toolchain $(CURDIR)/tmp/playwright $(CURDIR)/tmp/frontend-embed $(CURDIR)/.cache $(FRONTEND_NODE_MODULES_DIRS) $(CURDIR)/.pnpm-store
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

FORCE:

include tools/task_surface.generated.mk

$(SBOM_ARTIFACT) $(LICENSE_REPORT_ARTIFACT): $(NODE_BIN) $(FRONTEND_INSTALL_STAMP) $(CYCLONEDX_GOMOD_BIN) $(SYFT_BIN) scripts/generate-sbom-license-evidence.mjs scripts/validate-cyclonedx.mjs go.mod go.sum package.json pnpm-lock.yaml pnpm-workspace.yaml docker-compose.dev.yml $(wildcard apps/web/package.json packages/*/package.json)
	$(Q)mkdir -p $(RELEASE_ARTIFACT_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(Q)CARTULARY_TEST_TARGET=release-evidence CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_PHASE_SCRIPT) "generate SBOM/license evidence" -- env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack" GO="$(GO)" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" NODE_BIN="$(NODE_BIN)" PNPM="$(PNPM)" CYCLONEDX_GOMOD_BIN="$(CYCLONEDX_GOMOD_BIN)" SYFT_BIN="$(SYFT_BIN)" RELEASE_ARTIFACT_DIR="$(RELEASE_ARTIFACT_DIR)" LICENSE_REPORT_ARTIFACT="$(LICENSE_REPORT_ARTIFACT)" SBOM_ARTIFACT="$(SBOM_ARTIFACT)" $(NODE_BIN) ./scripts/generate-sbom-license-evidence.mjs

$(NODE_BIN): FORCE scripts/bootstrap-node-runtime.sh Makefile
	$(Q)NODE_VERSION="$(NODE_VERSION)" NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" ./scripts/bootstrap-node-runtime.sh

$(FRONTEND_TOOLCHAIN_STAMP): FORCE $(NODE_BIN) Makefile scripts/frontend-toolchain.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile frontend-toolchain --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/frontend-toolchain.sh --input scripts/cache-artifact.sh --output "$(FRONTEND_TOOLCHAIN_STAMP)" --key "node_version=$(NODE_VERSION)" --key "pnpm_version=$(PNPM_VERSION)" --key "node_bin=$(NODE_BIN)" --key "pnpm=$(PNPM)" -- env NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" NODE_BIN="$(NODE_BIN)" PNPM="$(PNPM)" NODE_VERSION="$(NODE_VERSION)" PNPM_VERSION="$(PNPM_VERSION)" FRONTEND_TOOLCHAIN_STAMP="$(FRONTEND_TOOLCHAIN_STAMP)" ./scripts/frontend-toolchain.sh

$(FRONTEND_INSTALL_STAMP): FORCE $(FRONTEND_INSTALL_INPUTS) $(FRONTEND_TOOLCHAIN_STAMP) scripts/frontend-install.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)CARTULARY_TEST_TARGET="frontend-install" $(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile frontend-install --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL $(foreach input,$(FRONTEND_INSTALL_INPUTS),--input "$(input)") --input .npmrc --input scripts/frontend-install.sh --input scripts/cache-artifact.sh --output "$(FRONTEND_INSTALL_STAMP)" --key "node_version=$(NODE_VERSION)" --key "pnpm_version=$(PNPM_VERSION)" --key "pnpm_install_flags=$(PNPM_INSTALL_FLAGS)" -- env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack" CI=true FRONTEND_INSTALL_STAMP="$(FRONTEND_INSTALL_STAMP)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" PNPM="$(PNPM)" PNPM_INSTALL_FLAGS="$(PNPM_INSTALL_FLAGS)" NODE_BIN="$(NODE_BIN)" NODE_VERSION="$(NODE_VERSION)" PNPM_VERSION="$(PNPM_VERSION)" bash ./scripts/frontend-install.sh

$(SQLC_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-sqlc --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(SQLC_BIN)" --key "tool=$(SQLC_TOOL)" --key "binary=sqlc" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(SQLC_BIN)" TOOL_MODULE="$(SQLC_TOOL)" TOOL_BINARY_NAME="sqlc" TOOL_LABEL="bootstrap sqlc tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(GOOSE_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-goose --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(GOOSE_BIN)" --key "tool=$(GOOSE_TOOL)" --key "binary=goose" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(GOOSE_BIN)" TOOL_MODULE="$(GOOSE_TOOL)" TOOL_BINARY_NAME="goose" TOOL_LABEL="bootstrap goose tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(STATICCHECK_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-staticcheck --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(STATICCHECK_BIN)" --key "tool=$(STATICCHECK_TOOL)" --key "binary=staticcheck" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(STATICCHECK_BIN)" TOOL_MODULE="$(STATICCHECK_TOOL)" TOOL_BINARY_NAME="staticcheck" TOOL_LABEL="bootstrap staticcheck tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(GOVULNCHECK_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-govulncheck --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(GOVULNCHECK_BIN)" --key "tool=$(GOVULNCHECK_TOOL)" --key "binary=govulncheck" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(GOVULNCHECK_BIN)" TOOL_MODULE="$(GOVULNCHECK_TOOL)" TOOL_BINARY_NAME="govulncheck" TOOL_LABEL="bootstrap govulncheck tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(GOSEC_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-gosec --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(GOSEC_BIN)" --key "tool=$(GOSEC_TOOL)" --key "binary=gosec" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(GOSEC_BIN)" TOOL_MODULE="$(GOSEC_TOOL)" TOOL_BINARY_NAME="gosec" TOOL_LABEL="bootstrap gosec tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(CYCLONEDX_GOMOD_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-cyclonedx-gomod --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(CYCLONEDX_GOMOD_BIN)" --key "tool=$(CYCLONEDX_GOMOD_TOOL)" --key "binary=cyclonedx-gomod" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(CYCLONEDX_GOMOD_BIN)" TOOL_MODULE="$(CYCLONEDX_GOMOD_TOOL)" TOOL_BINARY_NAME="cyclonedx-gomod" TOOL_LABEL="bootstrap cyclonedx-gomod tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(SYFT_BIN): FORCE Makefile scripts/bootstrap-go-tool.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile go-tool-syft --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-go-tool.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(SYFT_BIN)" --key "tool=$(SYFT_TOOL)" --key "binary=syft" -- env GO="$(GO)" TOOLBIN_DIR="$(TOOLBIN_DIR)" TOOL_OUTPUT="$(SYFT_BIN)" TOOL_MODULE="$(SYFT_TOOL)" TOOL_BINARY_NAME="syft" TOOL_LABEL="bootstrap syft tool" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/bootstrap-go-tool.sh

$(SHELLCHECK_BIN): FORCE scripts/bootstrap-shellcheck.sh Makefile $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile shellcheck-tool --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input Makefile --input scripts/bootstrap-shellcheck.sh --input scripts/cache-artifact.sh --output "$(SHELLCHECK_BIN)" --key "shellcheck_version=$(SHELLCHECK_VERSION)" -- env SHELLCHECK_VERSION="$(SHELLCHECK_VERSION)" TOOLBIN_DIR="$(TOOLBIN_DIR)" SHELLCHECK_BIN="$(SHELLCHECK_BIN)" CARTULARY_SHELLCHECK_ARCHIVE_DIR="$(SHELLCHECK_ARCHIVE_DIR)" ./scripts/bootstrap-shellcheck.sh

$(TEST_SERVICES_BIN): FORCE $$(TEST_SERVICES_BUILD_INPUTS) scripts/build-go-artifact.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.build_artifact.v1 --scope build-artifact --profile testservices-build --cache-dir "$(CARTULARY_BUILD_CACHE_DIR)" --disable-env CARTULARY_BUILD_CACHE_DISABLE --force-env CARTULARY_FORCE_REBUILD $(foreach input,$(TEST_SERVICES_BUILD_INPUTS),--input "$(input)") --input scripts/build-go-artifact.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(TEST_SERVICES_BIN)" --key "go=$${GO:-$(GO)}" --key "GOOS=$${GOOS:-}" --key "GOARCH=$${GOARCH:-}" --key "CGO_ENABLED=$${CGO_ENABLED:-}" --key "GOFLAGS=$${GOFLAGS:-}" -- env GO="$(GO)" BUILD_OUTPUT="$(TEST_SERVICES_BIN)" BUILD_PACKAGE="./tools/testservices" BUILD_LABEL="build testservices" CARTULARY_TEST_TARGET="testservices-build" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/build-go-artifact.sh

$(PLAYWRIGHT_INSTALL_STAMP): FORCE $(FRONTEND_INSTALL_STAMP) $(FRONTEND_TOOLCHAIN_STAMP) scripts/playwright-install.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.readiness.v1 --scope readiness --profile playwright-install --cache-dir "$(CARTULARY_READINESS_CACHE_DIR)" --disable-env CARTULARY_READINESS_DISABLE_CACHE --force-env CARTULARY_FORCE_REINSTALL --input package.json --input pnpm-lock.yaml --input apps/web/package.json --input scripts/playwright-install.sh --input scripts/cache-artifact.sh --output "$(PLAYWRIGHT_INSTALL_STAMP)" --key "node_version=$(NODE_VERSION)" --key "pnpm_version=$(PNPM_VERSION)" -- env PLAYWRIGHT_INSTALL_STAMP="$(PLAYWRIGHT_INSTALL_STAMP)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" NODE_BIN="$(NODE_BIN)" PNPM="$(PNPM)" NODE_VERSION="$(NODE_VERSION)" PNPM_VERSION="$(PNPM_VERSION)" ./scripts/playwright-install.sh

$(EMBEDDED_WEB_ASSET_STAMP) $(EMBEDDED_WEB_ASSET_ARCHIVE) $(EMBEDDED_WEB_ASSET_READY_STAMP) &: FORCE $(WEB_DIST_INDEX) scripts/embed-web-assets.sh tools/embedwebassets/main.go go.mod $(CACHE_ARTIFACT_SCRIPT)
	$(Q)$(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.build_artifact.v1 --scope build-artifact --profile embedded-web-assets --cache-dir "$(CARTULARY_BUILD_CACHE_DIR)" --disable-env CARTULARY_BUILD_CACHE_DISABLE --force-env CARTULARY_FORCE_REBUILD --input-dir "$(CURDIR)/apps/web/dist" --input scripts/embed-web-assets.sh --input tools/embedwebassets/main.go --input go.mod --input scripts/cache-artifact.sh --input "$(GO)" --output "$(EMBEDDED_WEB_ASSET_STAMP)" --output "$(EMBEDDED_WEB_ASSET_ARCHIVE)" --output "$(EMBEDDED_WEB_ASSET_READY_STAMP)" --output-dir "$(EMBEDDED_WEB_ASSET_DIR)" --key "source=$(WEB_DIST_INDEX)" --key "go=$${GO:-$(GO)}" -- env GO="$(GO)" WEB_DIST_INDEX="$(WEB_DIST_INDEX)" EMBEDDED_WEB_ASSET_DIR="$(EMBEDDED_WEB_ASSET_DIR)" EMBEDDED_WEB_ASSET_ARCHIVE="$(EMBEDDED_WEB_ASSET_ARCHIVE)" EMBEDDED_WEB_ASSET_STAMP="$(EMBEDDED_WEB_ASSET_STAMP)" EMBEDDED_WEB_ASSET_READY_STAMP="$(EMBEDDED_WEB_ASSET_READY_STAMP)" ./scripts/embed-web-assets.sh

$(SERVER_BIN): FORCE $$(SERVER_BUILD_INPUTS) $(EMBEDDED_WEB_ASSET_STAMP) $(EMBEDDED_WEB_ASSET_ARCHIVE) $(EMBEDDED_WEB_ASSET_READY_STAMP) scripts/build-go-artifact.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)CARTULARY_TEST_TARGET="build-server" $(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.build_artifact.v1 --scope build-artifact --profile build-server --cache-dir "$(CARTULARY_BUILD_CACHE_DIR)" --disable-env CARTULARY_BUILD_CACHE_DISABLE --force-env CARTULARY_FORCE_REBUILD $(foreach input,$(SERVER_BUILD_INPUTS),--input "$(input)") --input-dir "$(EMBEDDED_WEB_ASSET_DIR)" --input scripts/build-go-artifact.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(SERVER_BIN)" --key "go=$${GO:-$(GO)}" --key "GOOS=$${GOOS:-}" --key "GOARCH=$${GOARCH:-}" --key "CGO_ENABLED=$${CGO_ENABLED:-}" --key "GOFLAGS=$${GOFLAGS:-}" -- env GO="$(GO)" BUILD_OUTPUT="$(SERVER_BIN)" BUILD_PACKAGE="./cmd/server" BUILD_LABEL="build server" CARTULARY_TEST_TARGET="build-server" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/build-go-artifact.sh

$(MIGRATE_BIN): FORCE $$(MIGRATE_BUILD_INPUTS) scripts/build-go-artifact.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)CARTULARY_TEST_TARGET="build-migrate" $(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.build_artifact.v1 --scope build-artifact --profile build-migrate --cache-dir "$(CARTULARY_BUILD_CACHE_DIR)" --disable-env CARTULARY_BUILD_CACHE_DISABLE --force-env CARTULARY_FORCE_REBUILD $(foreach input,$(MIGRATE_BUILD_INPUTS),--input "$(input)") --input scripts/build-go-artifact.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(MIGRATE_BIN)" --key "go=$${GO:-$(GO)}" --key "GOOS=$${GOOS:-}" --key "GOARCH=$${GOARCH:-}" --key "CGO_ENABLED=$${CGO_ENABLED:-}" --key "GOFLAGS=$${GOFLAGS:-}" -- env GO="$(GO)" BUILD_OUTPUT="$(MIGRATE_BIN)" BUILD_PACKAGE="./cmd/migrate" BUILD_LABEL="build migrate" CARTULARY_TEST_TARGET="build-migrate" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/build-go-artifact.sh

$(OPERATOR_BIN): FORCE $$(OPERATOR_BUILD_INPUTS) scripts/build-go-artifact.sh $(CACHE_ARTIFACT_SCRIPT)
	$(Q)CARTULARY_TEST_TARGET="build-operator" $(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.build_artifact.v1 --scope build-artifact --profile build-operator --cache-dir "$(CARTULARY_BUILD_CACHE_DIR)" --disable-env CARTULARY_BUILD_CACHE_DISABLE --force-env CARTULARY_FORCE_REBUILD $(foreach input,$(OPERATOR_BUILD_INPUTS),--input "$(input)") --input scripts/build-go-artifact.sh --input scripts/cache-artifact.sh --input "$(GO)" --output "$(OPERATOR_BIN)" --key "go=$${GO:-$(GO)}" --key "GOOS=$${GOOS:-}" --key "GOARCH=$${GOARCH:-}" --key "CGO_ENABLED=$${CGO_ENABLED:-}" --key "GOFLAGS=$${GOFLAGS:-}" -- env GO="$(GO)" BUILD_OUTPUT="$(OPERATOR_BIN)" BUILD_PACKAGE="./cmd/operator" BUILD_LABEL="build operator" CARTULARY_TEST_TARGET="build-operator" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" ./scripts/build-go-artifact.sh

$(WEB_DIST_INDEX): FORCE $$(WEB_BUILD_INPUTS) $(FRONTEND_INSTALL_STAMP) scripts/build-web-artifact.sh $(CACHE_ARTIFACT_SCRIPT) | $(NODE_BIN)
	$(Q)CARTULARY_TEST_TARGET="build-web" $(CACHE_ARTIFACT_SCRIPT) --schema-id cartulary.cache.build_artifact.v1 --scope build-artifact --profile build-web --cache-dir "$(CARTULARY_BUILD_CACHE_DIR)" --disable-env CARTULARY_BUILD_CACHE_DISABLE --force-env CARTULARY_FORCE_REBUILD $(foreach input,$(WEB_BUILD_INPUTS),--input "$(input)") --input scripts/build-web-artifact.sh --input scripts/cache-artifact.sh --output-dir "$(CURDIR)/apps/web/dist" --key "node_version=$(NODE_VERSION)" --key "pnpm_version=$(PNPM_VERSION)" --key "vite_flags=$(VITE_BUILD_FLAGS)" -- env RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" PNPM="$(PNPM)" VITE_BUILD_FLAGS="$(VITE_BUILD_FLAGS)" CARTULARY_TEST_TARGET="build-web" ./scripts/build-web-artifact.sh
