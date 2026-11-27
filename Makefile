SHELL := /bin/bash
include tools/harness/workspace/admission.mk
include tools/workspace_layout.generated.mk
Q := @
# A read-only invocation needs only the projected command surface. Do not
# initialize writer identity, compiler discovery, caches, or build recipes.
ifeq ($(MAKECMDGOALS),test-run-status)
include tools/task_surface.runtime.generated.mk
include tools/task_surface.generated.mk
else
.DEFAULT_GOAL := help

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
RUN_STEP_SCRIPT := $(CURDIR)/tools/harness/execution/run-step.sh
TASK_SURFACE_MANIFEST ?= $(CURDIR)/tools/task_surface_manifest.json

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

include tools/task_surface.runtime.generated.mk
include tools/task_surface.generated.mk

endif # managed writer configuration
