#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GO_TARGET_HELPER="$ROOT_DIR/scripts/run-go-target.sh"
MANIFEST_HELPER="$ROOT_DIR/scripts/lib/phase-manifest.mjs"
node_bin="${NODE_BIN:-node}"

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output to omit [$needle]"
  fi
}

assert_not_zero() {
  local actual="$1"
  local label="$2"

  if [[ -z "$actual" || "$actual" == "0" ]]; then
    fail "$label: expected a non-zero value, got [$actual]"
  fi
}

backend_unit_core_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-unit backend-unit-core
)"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase0_" "backend-unit-core phase0 selector"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase2Unit_" "backend-unit-core phase2 selector"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase3Unit_" "backend-unit-core phase3 selector"

backend_unit_auth_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-unit backend-unit-auth
)"
assert_contains "$backend_unit_auth_shared_command" "TestSupportPhase1_" "backend-unit-auth phase1 selector"

phase0_platform_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-phase0-platform
)"
assert_contains "$phase0_platform_shared_command" "TestSupportPhase0_" "backend-integration phase0 platform support selector"
assert_contains "$phase0_platform_shared_command" "TestPhase0_SchemaBootstrap" "backend-integration phase0 platform authoritative selector"
assert_not_contains "$phase0_platform_shared_command" "TestPhase0_FirstAdminBootstrap" "backend-integration phase0 platform excludes app selector"

phase0_app_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration backend-integration-phase0-app
)"
assert_contains "$phase0_app_shared_command" "TestPhase0_FirstAdminBootstrap" "backend-integration phase0 app selector"
assert_not_contains "$phase0_app_shared_command" "TestSupportPhase0_" "backend-integration phase0 app excludes platform support selector"

phase2_incidents_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-phase2-incidents
)"
assert_contains "$phase2_incidents_shared_command" "TestSupportPhase2_" "backend-integration phase2 incidents support selector"
assert_contains "$phase2_incidents_shared_command" "TestPhase2_I_2_01" "backend-integration phase2 incidents authoritative selector"

phase2_incidents_shard_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration backend-integration-phase2-incidents-shard-02
)"
assert_contains "$phase2_incidents_shard_command" "TestPhase2_I_2_01" "backend-integration phase2 incidents planned shard selector"

phase2_incidents_support_shard_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-phase2-incidents-shard-04
)"
assert_contains "$phase2_incidents_support_shard_command" "TestSupportPhase2_" "backend-integration support phase2 planned shard selector"

phase3_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-phase3-timeline
)"
assert_contains "$phase3_timeline_shared_command" "TestSupportPhase3Integration_" "backend-integration phase3 timeline support selector"
assert_contains "$phase3_timeline_shared_command" "TestPhase3_I_3_01" "backend-integration phase3 timeline authoritative selector"

phase4_entities_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-phase4-entities
)"
assert_contains "$phase4_entities_shared_command" "TestSupportPhase4Integration_" "backend-integration phase4 entities support selector"
assert_contains "$phase4_entities_shared_command" "TestPhase4_ResolveRoute" "backend-integration phase4 entities authoritative selector"
assert_not_contains "$phase4_entities_shared_command" "TestPhase4_AutoResolutionEligibility" "backend-integration phase4 entities excludes timeline selector"

phase4_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration backend-integration-phase4-timeline
)"
assert_contains "$phase4_timeline_shared_command" "TestPhase4_AutoResolutionEligibility" "backend-integration phase4 timeline selector"
assert_not_contains "$phase4_timeline_shared_command" "TestSupportPhase4Integration_" "backend-integration phase4 timeline excludes entities support selector"

auth_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-auth
)"
assert_contains "$auth_shared_command" "TestSupportPhase1_" "backend-integration-auth phase1 selector"

backend_unit_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  assign_calls=0
  raw_calls=0
  support_calls=0
  manifest_calls=0
  phase4_manifest_calls=0
  raw_phase4_calls=0
  finish_status=""
  assign_captured_report() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    assign_calls=$((assign_calls + 1))
    dir_ref="/tmp/backend-unit-shared"
    usage_ref="actual"
  }
  emit_go_raw_phase() {
    local phase_label="$1"
    local regex="${4:-}"
    raw_calls=$((raw_calls + 1))
    if [[ "$phase_label" == *"phase4"* || "$regex" == *"U_4"* ]]; then
      raw_phase4_calls=$((raw_phase4_calls + 1))
    fi
  }
  emit_declared_support_phase() {
    support_calls=$((support_calls + 1))
  }
  emit_go_manifest_phase() {
    local phase_label="$1"
    manifest_calls=$((manifest_calls + 1))
    if [[ "$phase_label" == "backend-unit phase4 authoritative" ]]; then
      phase4_manifest_calls=$((phase4_manifest_calls + 1))
    fi
  }
  clear_go_selection_env() {
    :
  }
  finish_target() {
    finish_status="$1"
    printf "assign_calls=%s raw_calls=%s support_calls=%s manifest_calls=%s phase4_manifest_calls=%s raw_phase4_calls=%s finish_status=%s\n" "$assign_calls" "$raw_calls" "$support_calls" "$manifest_calls" "$phase4_manifest_calls" "$raw_phase4_calls" "$finish_status"
  }
  run_backend_unit
)"
assert_contains "$backend_unit_structure" "assign_calls=3" "backend-unit shared capture count"
assert_contains "$backend_unit_structure" "raw_calls=1" "backend-unit raw configtest count"
assert_contains "$backend_unit_structure" "support_calls=4" "backend-unit support phase count"
assert_contains "$backend_unit_structure" "phase4_manifest_calls=1" "backend-unit phase4 manifest phase count"
assert_contains "$backend_unit_structure" "raw_phase4_calls=0" "backend-unit raw phase4 phase count"
assert_contains "$backend_unit_structure" "finish_status=0" "backend-unit finish status"

backend_store_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  assign_calls=0
  raw_calls=0
  manifest_calls=0
  phase4_manifest_calls=0
  finish_status=""
  assign_captured_report() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    assign_calls=$((assign_calls + 1))
    dir_ref="/tmp/backend-store-shared"
    usage_ref="actual"
  }
  emit_go_raw_phase() {
    raw_calls=$((raw_calls + 1))
  }
  emit_go_manifest_phase() {
    local phase_label="$1"
    manifest_calls=$((manifest_calls + 1))
    if [[ "$phase_label" == "backend-store phase4 authoritative" ]]; then
      phase4_manifest_calls=$((phase4_manifest_calls + 1))
    fi
  }
  clear_go_selection_env() {
    :
  }
  finish_target() {
    finish_status="$1"
    printf "assign_calls=%s raw_calls=%s manifest_calls=%s phase4_manifest_calls=%s finish_status=%s\n" "$assign_calls" "$raw_calls" "$manifest_calls" "$phase4_manifest_calls" "$finish_status"
  }
  run_backend_store
)"
assert_contains "$backend_store_structure" "assign_calls=1" "backend-store single shared capture"
assert_contains "$backend_store_structure" "raw_calls=0" "backend-store raw phase count"
assert_contains "$backend_store_structure" "manifest_calls=4" "backend-store derived phase count"
assert_contains "$backend_store_structure" "phase4_manifest_calls=1" "backend-store phase4 manifest phase count"
assert_contains "$backend_store_structure" "finish_status=0" "backend-store finish status"

backend_integration_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  raw_calls=0
  manifest_calls=0
  phase4_manifest_calls=0
  finish_status=""
  capture_named_shared_reports_parallel() {
    capture_target="$1"
    capture_jobs="$2"
    local metadata_dir="$3"
    shift 3
    capture_calls=$((capture_calls + 1))
    mkdir -p "$metadata_dir"
    local shared_name
    for shared_name in "$@"; do
      capture_names="${capture_names} ${shared_name}"
      printf '/tmp/%s\nactual\n' "$shared_name" >"$metadata_dir/${shared_name}.meta"
    done
  }
  emit_go_raw_phase() {
    raw_calls=$((raw_calls + 1))
  }
  emit_go_manifest_phase() {
    local phase_label="$1"
    manifest_calls=$((manifest_calls + 1))
    if [[ "$phase_label" == backend-integration\ phase4\ authoritative* ]]; then
      phase4_manifest_calls=$((phase4_manifest_calls + 1))
    fi
  }
  create_aggregate_report() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    local aggregate_name="$4"
    dir_ref="/tmp/${aggregate_name}"
    usage_ref="actual"
  }
  clear_go_selection_env() {
    :
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s raw_calls=%s manifest_calls=%s phase4_manifest_calls=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$raw_calls" "$manifest_calls" "$phase4_manifest_calls" "$finish_status" "$capture_names"
  }
  run_backend_integration
)"
assert_contains "$backend_integration_structure" "capture_calls=1" "backend-integration parallel capture count"
assert_contains "$backend_integration_structure" "capture_target=backend-integration" "backend-integration parallel capture target"
assert_contains "$backend_integration_structure" "capture_jobs=4" "backend-integration default shard jobs"
assert_contains "$backend_integration_structure" "raw_calls=1" "backend-integration raw testutil phase count"
assert_contains "$backend_integration_structure" "manifest_calls=7" "backend-integration manifest shard phase count"
assert_contains "$backend_integration_structure" "phase4_manifest_calls=2" "backend-integration phase4 split phase count"
assert_contains "$backend_integration_structure" "backend-integration-phase4-entities-shard-01 backend-integration-phase4-entities-shard-02" "backend-integration captures split phase4 entity shards"
assert_contains "$backend_integration_structure" "backend-integration-phase2-incidents-shard-01" "backend-integration captures first split phase2 incident shard"
assert_contains "$backend_integration_structure" "backend-integration-phase2-incidents-shard-02" "backend-integration captures second split phase2 incident shard"
assert_contains "$backend_integration_structure" "backend-integration-testutil-shard-01" "backend-integration captures raw testutil shard"
assert_contains "$backend_integration_structure" "names= backend-integration-phase4-entities-shard-01" "backend-integration weighted shard order starts with heaviest shard"
assert_contains "$backend_integration_structure" "finish_status=0" "backend-integration finish status"

backend_integration_support_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  support_calls=0
  finish_status=""
  capture_named_shared_reports_parallel() {
    capture_target="$1"
    capture_jobs="$2"
    local metadata_dir="$3"
    shift 3
    capture_calls=$((capture_calls + 1))
    mkdir -p "$metadata_dir"
    local shared_name
    for shared_name in "$@"; do
      capture_names="${capture_names} ${shared_name}"
      printf '/tmp/%s\nactual\n' "$shared_name" >"$metadata_dir/${shared_name}.meta"
    done
  }
  emit_declared_support_phase() {
    support_calls=$((support_calls + 1))
  }
  create_aggregate_report() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    local aggregate_name="$4"
    dir_ref="/tmp/${aggregate_name}"
    usage_ref="actual"
  }
  clear_go_selection_env() {
    :
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s support_calls=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$support_calls" "$finish_status" "$capture_names"
  }
  run_backend_integration_support
)"
assert_contains "$backend_integration_support_structure" "capture_calls=1" "backend-integration-support parallel capture count"
assert_contains "$backend_integration_support_structure" "capture_target=backend-integration-support" "backend-integration-support parallel capture target"
assert_contains "$backend_integration_support_structure" "capture_jobs=4" "backend-integration-support default shard jobs"
assert_contains "$backend_integration_support_structure" "support_calls=5" "backend-integration-support support shard phase count"
assert_contains "$backend_integration_support_structure" "backend-integration-phase4-entities-shard-" "backend-integration-support captures phase4 entities shards"
assert_not_contains "$backend_integration_support_structure" "backend-integration-testutil" "backend-integration-support skips testutil shard"
assert_contains "$backend_integration_support_structure" "names= backend-integration-phase4-entities-shard-" "backend-integration-support weighted shard order starts with heaviest support shard"
assert_contains "$backend_integration_support_structure" "finish_status=0" "backend-integration-support finish status"

phase0_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase0 backend_unit ./internal/platform/... ./internal/app)"
phase1_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase1 backend_unit ./internal/modules/auth)"
phase2_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase2 backend_unit ./internal/modules/incidents)"
phase3_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase3 backend_unit ./internal/modules/timeline)"
phase1_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase1 backend_integration_support ./internal/modules/auth)"
phase2_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase2 backend_integration_support ./internal/modules/incidents)"
phase3_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase3 backend_integration_support ./internal/modules/timeline)"
phase4_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase4 backend_integration_support ./internal/modules/entities ./internal/modules/timeline)"
assert_not_zero "$phase0_backend_unit_support_count" "phase0 backend-unit support-go-count"
assert_not_zero "$phase1_backend_unit_support_count" "phase1 backend-unit support-go-count"
assert_not_zero "$phase2_backend_unit_support_count" "phase2 backend-unit support-go-count"
assert_not_zero "$phase3_backend_unit_support_count" "phase3 backend-unit support-go-count"
assert_not_zero "$phase1_support_count" "phase1 support-go-count"
assert_not_zero "$phase2_support_count" "phase2 support-go-count"
assert_not_zero "$phase3_support_count" "phase3 support-go-count"
assert_not_zero "$phase4_support_count" "phase4 support-go-count"
