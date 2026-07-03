#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GO_TARGET_RUNNER="$ROOT_DIR/scripts/cartulary-runner.mjs"
GO_TARGET_PLAN_COVERAGE_HELPER="$ROOT_DIR/scripts/check-go-target-plan-coverage.mjs"
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

NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_PLAN_COVERAGE_HELPER" --root "$ROOT_DIR" --commands --quiet

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

phase10_operator_pass_shard="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" --phase phase10 list-shards backend-process | grep 'scn-004-pass')"
phase10_operator_pass_shard_command="$(
  CARTULARY_GO_TARGET_PHASE=phase10 NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_RUNNER" go-target inspect-aggregate-command backend-process "$phase10_operator_pass_shard"
)"
assert_contains "$phase10_operator_pass_shard_command" "TestPhase10_E_10_01_ObjectStoreMigrationRunEmitsPassEvidence" "backend-process phase10 operator scenario shard selector"
assert_not_contains "$phase10_operator_pass_shard_command" "TestPhase10_E_10_01_ObjectStoreMigrationRunEmitsMismatchEvidence" "backend-process phase10 operator scenario shard excludes peer scenario"

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

backend_unit_aggregates="$("$node_bin" "$ROOT_DIR/tools/harness/planning/target-plan.mjs" list-aggregates backend-unit)"
assert_contains "$backend_unit_aggregates" "backend-unit-core" "backend-unit core aggregate"
assert_contains "$backend_unit_aggregates" "backend-unit-auth" "backend-unit auth aggregate"
assert_contains "$backend_unit_aggregates" "backend-unit-configtest" "backend-unit configtest aggregate"

backend_store_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-store)"
assert_contains "$backend_store_shards" "backend-store-shard-" "backend-store captures planned shards"
phase4_backend_store_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" --phase phase4 list-shards backend-store)"
assert_contains "$phase4_backend_store_shards" "phase4-backend-store-shard-" "phase-filtered backend-store shards carry phase prefix"
phase4_backend_store_first_shard="$(printf '%s\n' "$phase4_backend_store_shards" | head -n 1)"
phase4_backend_store_shard_target="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" --phase phase4 shard-field backend-store "$phase4_backend_store_first_shard" target)"
assert_contains "$phase4_backend_store_shard_target" "backend-store" "phase-filtered shard-field keeps shifted field argument"
phase4_backend_store_aggregate="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" --phase phase4 list-aggregates backend-store | head -n 1)"
phase4_backend_store_aggregate_phase="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" --phase phase4 aggregate-field backend-store "$phase4_backend_store_aggregate" phase)"
assert_contains "$phase4_backend_store_aggregate_phase" "phase4" "phase-filtered aggregate-field keeps shifted field argument"

backend_integration_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration)"
assert_contains "$backend_integration_shards" "backend-integration-entities-shard-" "backend-integration captures entity shards"
assert_contains "$backend_integration_shards" "$phase2_incidents_shard" "backend-integration captures planned phase2 incident shard"
assert_contains "$backend_integration_shards" "backend-integration-testutil-shard-01" "backend-integration captures raw testutil shard"
phase4_backend_integration_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" --phase phase4 list-shards backend-integration)"
assert_contains "$phase4_backend_integration_shards" "phase4-backend-integration-entities-shard-" "phase-filtered backend-integration captures phase4 entities shard"
assert_not_contains "$phase4_backend_integration_shards" "$phase2_incidents_shard" "phase-filtered backend-integration excludes phase2 shard"
first_backend_integration_shard="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration | head -n 1)"
assert_contains "$backend_integration_shards" "$first_backend_integration_shard" "backend-integration weighted shard order starts with heaviest shard"

backend_integration_support_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration-support)"
assert_contains "$backend_integration_support_shards" "backend-integration-entities-shard-" "backend-integration-support captures entities shards"
assert_not_contains "$backend_integration_support_shards" "backend-integration-testutil" "backend-integration-support skips testutil shard"
first_backend_integration_support_shard="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration-support | head -n 1)"
assert_contains "$backend_integration_support_shards" "$first_backend_integration_support_shard" "backend-integration-support weighted shard order starts with heaviest support shard"
