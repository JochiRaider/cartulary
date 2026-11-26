SHELL := /bin/bash
include tools/harness/workspace/admission.mk
ifeq ($(workspace_admitted),yes)
include tools/workspace_layout.generated.mk
Q := @
# A read-only invocation needs only the projected command surface. Do not
# initialize writer identity, compiler discovery, caches, or build recipes.
ifeq ($(MAKECMDGOALS),test-run-status)
include tools/task_surface.runtime.generated.mk
include tools/task_surface.generated.mk
else
.DEFAULT_GOAL := help

.SECONDEXPANSION:

GO ?= $(shell if command -v go >/dev/null 2>&1; then command -v go; elif [ -x /usr/local/go/bin/go ]; then printf /usr/local/go/bin/go; fi)
override GO_TOOLCHAIN := go1.27.4
GOFIPS140 ?= v1.0.0-c2097c7c
export GOFIPS140
ifeq ($(CARTULARY_PREPARATION_POLICY),installed_only)
override GOTOOLCHAIN := local
else
override GOTOOLCHAIN := $(GO_TOOLCHAIN)
endif
CONFIG_FILE ?= $(CURDIR)/configs/dev/config.toml
REVIEW_PROFILE ?= network_flow_claimed
CARTULARY_MACHINE_CACHE_DIR ?= $(if $(filter /%,$(strip $(XDG_CACHE_HOME))),$(strip $(XDG_CACHE_HOME))/cartulary,$(if $(filter /%,$(strip $(HOME))),$(strip $(HOME))/.cache/cartulary,))
CARTULARY_HARNESS_SCRATCH_ROOT ?= $(CARTULARY_MACHINE_CACHE_DIR)/harness-scratch
GO_CACHE_DIR ?= $(CARTULARY_MACHINE_CACHE_DIR)/go/build
GO_MOD_CACHE_DIR ?= $(CARTULARY_MACHINE_CACHE_DIR)/go/mod
GO_TMP_DIR ?= $(CARTULARY_MACHINE_CACHE_DIR)/go/tmp
NODE_VERSION ?= 24.15.0
PNPM_VERSION ?= 10.33.0
VITEST_MAX_WORKERS ?= 4
FIXTURE_THRESHOLD_MS ?= 30000
FIXTURE_TOP ?= 5
NODE_RUNTIME_DIR ?= $(CURDIR)/$(LAYOUT_NODE_RUNTIME)
NODE_BIN ?= $(NODE_RUNTIME_DIR)/bin/node
PNPM ?= $(NODE_RUNTIME_DIR)/bin/pnpm
REPORT_RENDERER_RECEIPT := $(CURDIR)/build/report-renderer/installation.json
GOVULNCHECK_FLAGS ?= -test
GOVULNCHECK_PATTERNS ?= ./cmd/... ./internal/... ./db/... ./tools/...
GOVULNCHECK_DB ?=
GOSEC_RULES ?= G602,G124,G112,G114
GOSEC_FLAGS ?= -exclude-generated
GOSEC_PATTERNS ?= ./cmd/... ./internal/... ./db/... ./tools/...
GOSEC_TARGETED_RUNTIME_RULES ?= G122,G301,G302,G303,G304,G305,G306,G307
GOSEC_TARGETED_RUNTIME_FLAGS ?=
GOSEC_TARGETED_RUNTIME_PATTERNS ?= ./cmd/... ./internal/...
GOSEC_AUDIT_RUNTIME_RULES ?= G118,G122,G301,G302,G303,G304,G305,G306,G307
GOSEC_AUDIT_RUNTIME_FLAGS ?=
GOSEC_AUDIT_RUNTIME_PATTERNS ?= ./cmd/... ./internal/...
GOSEC_AUDIT_SUPPORT_RULES ?= G122,G301,G302,G303,G304,G305,G306,G307
GOSEC_AUDIT_SUPPORT_FLAGS ?= -exclude-generated -no-fail -quiet
GOSEC_AUDIT_SUPPORT_PATTERNS ?=
SCHEDULER_MANIFEST ?= $(CURDIR)/tools/scheduler_manifest.json
EXECUTION_TOPOLOGY_MANIFEST ?= $(CURDIR)/tools/execution_topology_manifest.json
OBJECT_STORE_BUCKET ?= cartulary
CARTULARY_BUILD_CACHE_DIR ?= $(CURDIR)/$(LAYOUT_BUILD_ARTIFACTS)
RUN_STEP_SCRIPT := $(CURDIR)/tools/harness/execution/run-step.sh
TEST_OUTPUT_SCRIPT := $(CURDIR)/tools/harness/output/test-output.mjs
CARTULARY_RUNNER_SCRIPT := $(CURDIR)/tools/harness/execution/cartulary-runner-cli.mjs
TASK_SURFACE_MANIFEST ?= $(CURDIR)/tools/task_surface_manifest.json
GO_TOOLCHAIN_READINESS_IMPL := tools/harness/readiness/go-toolchain-readiness.sh
GO_TOOLCHAIN_READINESS = bash $(GO_TOOLCHAIN_READINESS_IMPL) ensure
HARNESS_CONTRACT_CLI := $(CURDIR)/tools/harness/contract/harness-contract-cli.mjs
FOUNDATION_SCHEMA_VALIDATORS := $(CURDIR)/tools/harness/contract/generated/foundation-schema-validators.cjs
RUN_HARNESS_PREFLIGHT = $(NODE_BIN) $(HARNESS_CONTRACT_CLI) preflight

DEFAULT_CARTULARY_TEST_RESULTS_DIR := $(CURDIR)/$(LAYOUT_TEST_RESULTS)
CARTULARY_OUTPUT_MODE ?=
CARTULARY_TEST_RESULTS_DIR ?= $(DEFAULT_CARTULARY_TEST_RESULTS_DIR)
# One invocation has one identity, including every exported child environment.
ifeq ($(origin CARTULARY_TEST_RUN_ID),undefined)
CARTULARY_TEST_RUN_ID := $(shell if [ -x /usr/bin/date ]; then now="$$(/usr/bin/date -u +%Y%m%dT%H%M%SZ)"; elif command -v date >/dev/null 2>&1; then now="$$(date -u +%Y%m%dT%H%M%SZ)"; else now="unknown-time"; fi; printf '%s-p%s' "$$now" "$$$$")
endif
RELEASE_ARTIFACT_DIR ?= $(CURDIR)/$(LAYOUT_RELEASE_ARTIFACTS)
LICENSE_REPORT_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/license-report.json
SBOM_ARTIFACT ?= $(RELEASE_ARTIFACT_DIR)/sbom.cyclonedx.json
BENCHMARK_MANIFEST ?= $(CURDIR)/.cartulary/benchmark/benchmark_manifest.json
export GO CARTULARY_OUTPUT_MODE VERBOSE CI_VERBOSE CI CARTULARY_TEST_RESULTS_DIR CARTULARY_TEST_RUN_ID CARTULARY_TEST_INVENTORY CARTULARY_MACHINE_CACHE_DIR CARTULARY_HARNESS_SCRATCH_ROOT GO_CACHE_DIR GO_MOD_CACHE_DIR GO_TMP_DIR GOTOOLCHAIN GO_TOOLCHAIN

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
BIOME_CHECK_FLAGS := --reporter=summary --diagnostic-level=warn
export NO_COLOR := 1
export CLICOLOR := 0
export FORCE_COLOR := 0
endif

TESTCONTAINERS_GO_VERSION := v0.42.0

# PRODUCE runs producers registered in tools/harness_producers.json through
# the harness artifact store (P5-3). `go run` needs no prebuilt tool, so it
# also builds the harness runtime and installs the Node runtime. Each
# producer publishes at its binding, which the generated include defaults and
# exports.
PRODUCE = mkdir -p "$(GO_CACHE_DIR)" "$(GO_MOD_CACHE_DIR)" "$(GO_TMP_DIR)" && env GOCACHE="$(GO_CACHE_DIR)" GOMODCACHE="$(GO_MOD_CACHE_DIR)" GOTMPDIR="$(GO_TMP_DIR)" NODE_BIN="$(NODE_BIN)" NODE_RUNTIME_DIR="$(NODE_RUNTIME_DIR)" PNPM="$(PNPM)" GO="$(GO)" RUN_STEP_SCRIPT="$(RUN_STEP_SCRIPT)" "$(GO)" run ./tools/harnessbuild produce

FORCE:

include tools/task_surface.runtime.generated.mk
include tools/task_surface.generated.mk

$(SBOM_ARTIFACT) $(LICENSE_REPORT_ARTIFACT) &: tools/release-evidence/generate-sbom-license-evidence.mjs tools/release-evidence/validate-cyclonedx.mjs tools/release-evidence/validate-license-report.mjs tools/release-evidence/validate-release-sbom.mjs tools/schemas/cartulary.license_report.v2.schema.json go.mod go.sum package.json pnpm-lock.yaml pnpm-workspace.yaml docker-compose.dev.yml $(wildcard apps/web/package.json packages/*/package.json)
	$(Q)$(GO_TOOLCHAIN_READINESS)
	$(Q)$(PRODUCE) frontend-install cyclonedx-gomod syft
	$(Q)mkdir -p $(RELEASE_ARTIFACT_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR) $(GO_TMP_DIR)
	$(Q)CARTULARY_TEST_TARGET=release-evidence CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_STEP_SCRIPT) "generate SBOM/license evidence" -- env PATH="$(NODE_RUNTIME_DIR)/bin:$$PATH" COREPACK_HOME="$(NODE_RUNTIME_DIR)/corepack" GO="$(GO)" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" GO_TMP_DIR="$(GO_TMP_DIR)" NODE_BIN="$(NODE_BIN)" PNPM="$(PNPM)" CYCLONEDX_GOMOD_BIN="$(CYCLONEDX_GOMOD_BIN)" SYFT_BIN="$(SYFT_BIN)" RELEASE_ARTIFACT_DIR="$(RELEASE_ARTIFACT_DIR)" LICENSE_REPORT_ARTIFACT="$(LICENSE_REPORT_ARTIFACT)" SBOM_ARTIFACT="$(SBOM_ARTIFACT)" $(NODE_BIN) ./tools/release-evidence/generate-sbom-license-evidence.mjs

$(FOUNDATION_SCHEMA_VALIDATORS): internal/testutil/harnessdiag/testdata/schema-probes.json tools/task_surface_owner.json tools/harness/generated-artifacts/native-command-policy.mjs tools/harness/generated-artifacts/native-schema-validators.mjs tools/harness/generated-artifacts/schema-validator-compiler.mjs tools/harness/generated-artifacts/generate-foundation-schema-validators.mjs tools/harness/generated-artifacts/generated-transaction.mjs tools/harness/generated-artifacts/instrumentation-definitions.mjs tools/harness/generated-artifacts/instrumentation-shapes.mjs tools/harness/generated-artifacts/admission-vocabulary.mjs tools/harness/generated-artifacts/lifecycle-policy.mjs tools/harness/generated-artifacts/native-runtime-policy.mjs tools/harness/generated-artifacts/redaction-policy.mjs tools/harness/contract/failure-taxonomy.mjs tools/harness/contract/redaction-vocabulary.mjs tools/harness_instrumentation_definitions.json tools/harness_lifecycle_policy.json tools/harness_redaction_manifest.json tools/harness_schema_attachments.json tools/scheduler_resource_registry.json $(wildcard tools/schemas/*.schema.json)
	$(Q)$(PRODUCE) node-runtime
	$(Q)$(NODE_BIN) ./tools/harness/generated-artifacts/generate-foundation-schema-validators.mjs

$(REPORT_RENDERER_RECEIPT): FORCE tools/report-renderer/build-runtime.mjs
	$(Q)$(GO_TOOLCHAIN_READINESS)
	$(Q)$(PRODUCE) playwright
	$(Q)CARTULARY_TEST_TARGET=build-report-renderer CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(RUN_STEP_SCRIPT) "build isolated reporting runtime" -- env GO="$(GO)" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" GO_TMP_DIR="$(GO_TMP_DIR)" NODE_BIN="$(NODE_BIN)" PNPM="$(PNPM)" RUN_STEP_SCRIPT="$(RUN_STEP_SCRIPT)" $(NODE_BIN) tools/report-renderer/build-runtime.mjs

endif # workspace admission

endif # managed writer configuration
