#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GO_TARGET_RUNNER="$ROOT_DIR/scripts/cartulary-runner.mjs"
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

find_planned_shard_for_symbol() {
  local target="$1"
  local symbol="$2"

  "$node_bin" - "$ROOT_DIR" "$target" "$symbol" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root, target, symbol] = process.argv.slice(2);
const plan = JSON.parse(execFileSync(process.execPath, [path.join(root, "scripts/print-go-shard-plan.mjs"), "--json", "--target", target], { encoding: "utf8", cwd: root }));
const shard = plan.shards.find((candidate) => candidate.items.some((item) => item.symbol === symbol));
if (!shard) {
  process.exit(1);
}
process.stdout.write(shard.name);
EOF
}

backend_unit_core_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-unit backend-unit-core
)"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase0_" "backend-unit-core phase0 selector"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase2Unit_" "backend-unit-core phase2 selector"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase3Unit_" "backend-unit-core phase3 selector"

backend_unit_auth_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-unit backend-unit-auth
)"
assert_contains "$backend_unit_auth_shared_command" "TestSupportPhase1_" "backend-unit-auth phase1 selector"

phase0_platform_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration-support backend-integration-platform
)"
assert_contains "$phase0_platform_shared_command" "TestSupportPhase0_" "backend-integration phase0 platform support selector"
phase0_platform_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration backend-integration-platform
)"
assert_contains "$phase0_platform_authoritative_command" "TestPhase0_SchemaBootstrap" "backend-integration phase0 platform authoritative selector"
assert_not_contains "$phase0_platform_authoritative_command" "TestPhase0_FirstAdminBootstrap" "backend-integration phase0 platform excludes app selector"

phase0_app_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration backend-integration-app
)"
assert_contains "$phase0_app_shared_command" "TestPhase0_FirstAdminBootstrap" "backend-integration phase0 app selector"
assert_not_contains "$phase0_app_shared_command" "TestSupportPhase0_" "backend-integration phase0 app excludes platform support selector"

phase2_incidents_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration-support backend-integration-incidents
)"
assert_contains "$phase2_incidents_shared_command" "TestSupportPhase2_" "backend-integration phase2 incidents support selector"
phase2_incidents_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration backend-integration-incidents
)"
assert_contains "$phase2_incidents_authoritative_command" "TestPhase2_I_2_01" "backend-integration phase2 incidents authoritative selector"

phase2_incidents_shard="$(find_planned_shard_for_symbol backend-integration TestPhase2_I_2_01_IncidentCreatePersistsBootstrapStateAndRollsBackAtomically)"
phase2_incidents_shard_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration "$phase2_incidents_shard"
)"
assert_contains "$phase2_incidents_shard_command" "TestPhase2_I_2_01" "backend-integration phase2 incidents planned shard selector"

phase2_incidents_support_shard="$(find_planned_shard_for_symbol backend-integration-support TestSupportPhase2_ControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied)"
phase2_incidents_support_shard_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration-support "$phase2_incidents_support_shard"
)"
assert_contains "$phase2_incidents_support_shard_command" "TestSupportPhase2_" "backend-integration support phase2 planned shard selector"

phase3_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration-support backend-integration-timeline
)"
assert_contains "$phase3_timeline_shared_command" "TestSupportPhase3Integration_" "backend-integration phase3 timeline support selector"
phase3_timeline_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration backend-integration-timeline
)"
assert_contains "$phase3_timeline_authoritative_command" "TestPhase3_I_3_01" "backend-integration phase3 timeline authoritative selector"

phase4_entities_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration-support backend-integration-entities
)"
assert_contains "$phase4_entities_shared_command" "TestSupportPhase4Integration_" "backend-integration phase4 entities support selector"
phase4_entities_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration backend-integration-entities
)"
assert_contains "$phase4_entities_authoritative_command" "TestPhase4_ResolveRoute" "backend-integration phase4 entities authoritative selector"
assert_not_contains "$phase4_entities_authoritative_command" "TestPhase4_AutoResolutionEligibility" "backend-integration phase4 entities excludes timeline selector"

phase4_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration backend-integration-timeline
)"
assert_contains "$phase4_timeline_shared_command" "TestPhase4_AutoResolutionEligibility" "backend-integration phase4 timeline selector"
assert_not_contains "$phase4_timeline_shared_command" "TestSupportPhase4Integration_" "backend-integration phase4 timeline excludes entities support selector"

auth_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-integration-support backend-integration-auth
)"
assert_contains "$auth_shared_command" "TestSupportPhase1_" "backend-integration-auth phase1 selector"

backend_unit_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  assign_calls=0
  emit_calls=0
  assigned_names=""
  emitted_names=""
  finish_status=""
  assign_execution_family() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    local family="$4"
    assign_calls=$((assign_calls + 1))
    assigned_names="${assigned_names} ${family}"
    dir_ref="/tmp/${family}"
    usage_ref="actual"
  }
  emit_execution_family() {
    emit_calls=$((emit_calls + 1))
    emitted_names="${emitted_names} $2"
  }
  finish_target() {
    finish_status="$1"
    printf "assign_calls=%s emit_calls=%s finish_status=%s assigned=%s emitted=%s\n" "$assign_calls" "$emit_calls" "$finish_status" "$assigned_names" "$emitted_names"
  }
  run_unsharded_target backend-unit
)"
assert_contains "$backend_unit_structure" "assign_calls=3" "backend-unit aggregate capture count"
assert_contains "$backend_unit_structure" "emit_calls=3" "backend-unit aggregate emission count"
assert_contains "$backend_unit_structure" "backend-unit-core" "backend-unit core aggregate"
assert_contains "$backend_unit_structure" "backend-unit-auth" "backend-unit auth aggregate"
assert_contains "$backend_unit_structure" "backend-unit-configtest" "backend-unit configtest aggregate"
assert_contains "$backend_unit_structure" "finish_status=0" "backend-unit finish status"

backend_store_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  finalize_calls=0
  finalize_target=""
  finish_status=""
  capture_named_shared_reports_parallel() {
    capture_target="$1"
    capture_jobs="$2"
    shift 3
    capture_calls=$((capture_calls + 1))
    capture_names="$*"
  }
  finalize_scheduled_shards() {
    finalize_calls=$((finalize_calls + 1))
    finalize_target="$1"
    finish_target 0
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s finalize_calls=%s finalize_target=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$finalize_calls" "$finalize_target" "$finish_status" "$capture_names"
  }
  run_sharded_target backend-store
)"
assert_contains "$backend_store_structure" "capture_calls=1" "backend-store shard capture count"
assert_contains "$backend_store_structure" "capture_target=backend-store" "backend-store shard capture target"
assert_contains "$backend_store_structure" "capture_jobs=4" "backend-store default shard jobs"
assert_contains "$backend_store_structure" "finalize_calls=1" "backend-store finalize count"
assert_contains "$backend_store_structure" "backend-store-shard-" "backend-store captures planned shards"
assert_contains "$backend_store_structure" "finish_status=0" "backend-store finish status"

backend_integration_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  finalize_calls=0
  finalize_target=""
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
  finalize_scheduled_shards() {
    finalize_calls=$((finalize_calls + 1))
    finalize_target="$1"
    finish_target 0
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s finalize_calls=%s finalize_target=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$finalize_calls" "$finalize_target" "$finish_status" "$capture_names"
  }
  run_sharded_target backend-integration
)"
assert_contains "$backend_integration_structure" "capture_calls=1" "backend-integration parallel capture count"
assert_contains "$backend_integration_structure" "capture_target=backend-integration" "backend-integration parallel capture target"
assert_contains "$backend_integration_structure" "capture_jobs=4" "backend-integration default shard jobs"
assert_contains "$backend_integration_structure" "finalize_calls=1" "backend-integration finalize count"
assert_contains "$backend_integration_structure" "backend-integration-entities-shard-" "backend-integration captures entity shards"
assert_contains "$backend_integration_structure" "$phase2_incidents_shard" "backend-integration captures planned phase2 incident shard"
assert_contains "$backend_integration_structure" "backend-integration-testutil-shard-01" "backend-integration captures raw testutil shard"
first_backend_integration_shard="$("$node_bin" "$ROOT_DIR/scripts/lib/go-shard-plan.mjs" list-shards backend-integration | head -n 1)"
assert_contains "$backend_integration_structure" "names= $first_backend_integration_shard" "backend-integration weighted shard order starts with heaviest shard"
assert_contains "$backend_integration_structure" "finish_status=0" "backend-integration finish status"

backend_integration_support_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  finalize_calls=0
  finalize_target=""
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
  finalize_scheduled_shards() {
    finalize_calls=$((finalize_calls + 1))
    finalize_target="$1"
    finish_target 0
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s finalize_calls=%s finalize_target=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$finalize_calls" "$finalize_target" "$finish_status" "$capture_names"
  }
  run_sharded_target backend-integration-support
)"
assert_contains "$backend_integration_support_structure" "capture_calls=1" "backend-integration-support parallel capture count"
assert_contains "$backend_integration_support_structure" "capture_target=backend-integration-support" "backend-integration-support parallel capture target"
assert_contains "$backend_integration_support_structure" "capture_jobs=4" "backend-integration-support default shard jobs"
assert_contains "$backend_integration_support_structure" "finalize_calls=1" "backend-integration-support finalize count"
assert_contains "$backend_integration_support_structure" "backend-integration-entities-shard-" "backend-integration-support captures entities shards"
assert_not_contains "$backend_integration_support_structure" "backend-integration-testutil" "backend-integration-support skips testutil shard"
first_backend_integration_support_shard="$("$node_bin" "$ROOT_DIR/scripts/lib/go-shard-plan.mjs" list-shards backend-integration-support | head -n 1)"
assert_contains "$backend_integration_support_structure" "names= $first_backend_integration_support_shard" "backend-integration-support weighted shard order starts with heaviest support shard"
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
