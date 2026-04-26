#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_HELPER="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
PLAN_SCRIPT="$ROOT_DIR/scripts/print-target-plan.mjs"
SHARD_PLAN_SCRIPT="$ROOT_DIR/scripts/print-go-shard-plan.mjs"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/target-plan-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")

json_a="$tmp_dir/target-plan-a.json"
json_b="$tmp_dir/target-plan-b.json"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_a"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_b"

"$NODE_HELPER" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$json_a"
cmp -s "$json_a" "$json_b" || fail "target-plan JSON must be deterministic across invocations"

if ! "$NODE_HELPER" - "$json_a" <<'EOF'
const fs = require("node:fs");
const rows = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const storeRows = rows.filter((row) => row.target === "backend-store");
if (storeRows.length === 0 || !storeRows.every((row) => row.fixture_policy?.postgres === "transaction")) {
  process.exit(1);
}
const rawPgtest = rows.find((row) => row.target === "backend-integration" && row.shared_report === "backend-integration-testutil");
if (!rawPgtest || rawPgtest.fixture_policy?.postgres !== "package_reset") {
  process.exit(1);
}
const serviceBackedGoRows = rows.filter((row) => row.service_backed && row.runner_family === "go_test");
const validPolicies = new Set(["template_clone", "package_reset", "migration_scratch", "transaction", "group_clone"]);
if (serviceBackedGoRows.length === 0 || !serviceBackedGoRows.every((row) => validPolicies.has(row.fixture_policy?.postgres))) {
  process.exit(1);
}
const packageResetRows = serviceBackedGoRows.filter((row) => row.fixture_policy?.postgres === "package_reset" && row.coverage !== "raw");
if (!packageResetRows.every((row) => Number.isInteger(row.fixture_budget?.postgres?.max_package_resets) && Number.isInteger(row.fixture_budget?.postgres?.max_reset_duration_ms))) {
  process.exit(1);
}
const transactionRows = serviceBackedGoRows.filter((row) => row.fixture_policy?.postgres === "transaction");
if (!transactionRows.every((row) => Number.isInteger(row.fixture_budget?.postgres?.max_transactions))) {
  process.exit(1);
}
EOF
then
  fail "target-plan JSON must expose postgres fixture policies"
fi

shard_json_a="$tmp_dir/go-shard-plan-a.json"
shard_json_b="$tmp_dir/go-shard-plan-b.json"
"$NODE_HELPER" "$SHARD_PLAN_SCRIPT" --json --target backend-integration >"$shard_json_a"
"$NODE_HELPER" "$SHARD_PLAN_SCRIPT" --json --target backend-integration >"$shard_json_b"
"$NODE_HELPER" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$shard_json_a"
cmp -s "$shard_json_a" "$shard_json_b" || fail "go shard-plan JSON must be deterministic across invocations"

if ! "$NODE_HELPER" - "$shard_json_a" <<'EOF'
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const weights = plan.shards.map((shard) => shard.weight_ms);
for (let index = 1; index < weights.length; index += 1) {
  if (weights[index - 1] < weights[index]) {
    process.exit(1);
  }
}
const incidents = plan.aggregates.find((aggregate) => aggregate.name === "backend-integration-phase2-incidents");
if (!incidents || incidents.shards.length < 2) {
  process.exit(1);
}
if (!plan.shards.every((shard) => Number.isInteger(shard.shard_target_ms) && shard.shard_target_ms > 0)) {
  process.exit(1);
}
const integrationMultiItemShards = plan.shards.filter((shard) => shard.shard_target_ms === 18000 && (shard.has_authoritative || shard.has_support) && shard.item_count > 1);
if (!integrationMultiItemShards.every((shard) => shard.weight_ms <= 18000 && shard.shard_target_ms === 18000)) {
  process.exit(1);
}
const badIncidentShard = plan.shards.find((shard) =>
  shard.aggregate_name === "backend-integration-phase2-incidents" &&
  shard.items.some((item) => item.symbol.includes("ControlBoundary") && item.kind === "authoritative") &&
  shard.items.some((item) => item.symbol.includes("ControlBoundary") && item.kind === "support")
);
if (badIncidentShard) {
  process.exit(1);
}
const authoritative = plan.shards.flatMap((shard) => shard.items).filter((item) => item.kind === "authoritative");
const validPolicies = new Set(["template_clone", "package_reset", "migration_scratch", "transaction", "group_clone"]);
if (authoritative.length === 0 || !authoritative.every((item) => validPolicies.has(item.postgres_fixture_policy))) {
  process.exit(1);
}
const packageReset = authoritative.filter((item) => item.postgres_fixture_policy === "package_reset");
if (!packageReset.every((item) => Number.isInteger(item.postgres_fixture_budget?.max_package_resets) && Number.isInteger(item.postgres_fixture_budget?.max_reset_duration_ms))) {
  process.exit(1);
}
const shared = plan.shards.filter((shard) => shard.shared_across_targets);
if (shared.length === 0) {
  process.exit(1);
}
if (!plan.shards.every((shard) => shard.shared_across_targets === (shard.has_authoritative && shard.has_support))) {
  process.exit(1);
}
EOF
then
  fail "backend-integration go shard plan must be weighted, policy-bearing, split heavy aggregates, and mark cross-target shared shards"
fi

backend_store_shard_json="$tmp_dir/go-shard-plan-backend-store.json"
"$NODE_HELPER" "$SHARD_PLAN_SCRIPT" --json --target backend-store >"$backend_store_shard_json"
if ! "$NODE_HELPER" - "$backend_store_shard_json" <<'EOF'
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (!plan.targets.includes("backend-store") || plan.shards.length === 0) {
  process.exit(1);
}
const items = plan.shards.flatMap((shard) => shard.items);
if (items.length === 0 || !items.every((item) => item.kind === "authoritative" && item.postgres_fixture_policy === "transaction")) {
  process.exit(1);
}
EOF
then
  fail "backend-store go shard plan must expose authoritative transaction fixture planning"
fi

backend_store_output="$("$NODE_HELPER" "$PLAN_SCRIPT" --target backend-store)"
assert_contains "$backend_store_output" "backend-store service_backed=1" "backend-store compact target plan"
assert_contains "$backend_store_output" "rows=" "backend-store compact row count"
assert_contains "$backend_store_output" "shared_reports=" "backend-store compact shared report count"

default_output="$("$NODE_HELPER" "$PLAN_SCRIPT")"
detail_output="$("$NODE_HELPER" "$PLAN_SCRIPT" --detail)"
default_lines="$(printf '%s\n' "$default_output" | wc -l | tr -d '[:space:]')"
detail_lines="$(printf '%s\n' "$detail_output" | wc -l | tr -d '[:space:]')"
if (( default_lines >= detail_lines )); then
  fail "default target-plan output must be more compact than --detail"
fi

backend_store_detail="$("$NODE_HELPER" "$PLAN_SCRIPT" --detail --target backend-store)"
for phase in phase1 phase2 phase3 phase4; do
  assert_contains "$backend_store_detail" "$phase unit authoritative" "backend-store detailed target plan"
done
assert_contains "$backend_store_detail" "packages:" "backend-store detail packages"

results_dir="$tmp_dir/results"
make_output="$(
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store
)"
assert_contains "$make_output" "backend-store" "make explain-target"
assert_contains "$make_output" "phase1 unit authoritative" "make explain-target detailed default"
if [[ -d "$results_dir" ]] && [[ -n "$(find "$results_dir" -mindepth 1 -print -quit)" ]]; then
  fail "make explain-target must not create test report artifacts"
fi

make_compact_output="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=0)"
assert_contains "$make_compact_output" "backend-store service_backed=1" "make explain-target compact mode"

phase_root="$tmp_dir/phase-root"
mkdir -p "$phase_root/tools"
cp "$ROOT_DIR"/tools/phase*_test_map.json "$phase_root/tools/"
cat >"$phase_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [],
  "integration": [],
  "e2e": [],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_support_test.go",
      "selection_pattern": "TestSupportPhase5_",
      "symbol": "TestSupportPhase5_Discovered"
    }
  ]
}
JSON

discovered_phases="$(CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" list-phases)"
assert_contains "$discovered_phases" "phase5" "phase manifest discovery includes phase5"

phase5_plan="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
    "$NODE_HELPER" "$PLAN_SCRIPT" --json
)"
assert_contains "$phase5_plan" '"manifest_phase": "phase5"' "target-plan support rows include discovered phase"

phase5_shared_command="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
  NODE_BIN="$NODE_HELPER" \
    "$ROOT_DIR/scripts/run-go-target.sh" inspect-shared-command backend-unit backend-unit-auth
)"
assert_contains "$phase5_shared_command" "TestSupportPhase5_Discovered" "run-go-target support selection includes discovered phase"

invalid_phase_root="$tmp_dir/invalid-phase-root"
mkdir -p "$invalid_phase_root/tools"
cat >"$invalid_phase_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase5_Invalid_U_5_01",
      "execution_dependency": "backend_store",
      "evidence_layer": "store_domain",
      "fixture_policy": { "postgres": "invalid" },
      "claim": "invalid fixture policy smoke",
      "out_of_scope": "invalid fixture policy smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$invalid_phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase5 >/dev/null 2>&1; then
  fail "phase manifest validation must reject unknown postgres fixture policies"
fi

missing_policy_root="$tmp_dir/missing-policy-root"
mkdir -p "$missing_policy_root/tools"
cat >"$missing_policy_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase5_MissingPolicy_U_5_01",
      "execution_dependency": "backend_store",
      "evidence_layer": "store_domain",
      "claim": "missing fixture policy smoke",
      "out_of_scope": "missing fixture policy smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$missing_policy_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase5 >/dev/null 2>&1; then
  fail "phase manifest validation must reject missing service-backed postgres fixture policies"
fi

missing_budget_root="$tmp_dir/missing-budget-root"
mkdir -p "$missing_budget_root/tools"
cat >"$missing_budget_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase5_MissingBudget_U_5_01",
      "execution_dependency": "backend_store",
      "evidence_layer": "store_domain",
      "fixture_policy": { "postgres": "package_reset" },
      "claim": "missing fixture budget smoke",
      "out_of_scope": "missing fixture budget smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$missing_budget_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase5 >/dev/null 2>&1; then
  fail "phase manifest validation must reject package_reset without postgres fixture budgets"
fi

invalid_budget_root="$tmp_dir/invalid-budget-root"
mkdir -p "$invalid_budget_root/tools"
cat >"$invalid_budget_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase5_InvalidBudget_U_5_01",
      "execution_dependency": "backend_store",
      "evidence_layer": "store_domain",
      "fixture_policy": { "postgres": "package_reset" },
      "fixture_budget": {
        "postgres": {
          "max_package_resets": 1,
          "max_reset_duration_ms": 10000,
          "dirty_tables": ["users", "users"]
        }
      },
      "claim": "invalid fixture budget smoke",
      "out_of_scope": "invalid fixture budget smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$invalid_budget_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase5 >/dev/null 2>&1; then
  fail "phase manifest validation must reject invalid postgres fixture budgets"
fi

missing_migration_reason_root="$tmp_dir/missing-migration-reason-root"
mkdir -p "$missing_migration_reason_root/tools"
cat >"$missing_migration_reason_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "support_go_targets": [
    {
      "target": "backend_integration_support",
      "section": "integration",
      "package": "./internal/platform/objectstore",
      "file": "internal/platform/objectstore/objectstore_phase0_support_test.go",
      "symbol": "TestSupportPhase0_ManagedServiceObjectStoreBinding",
      "selection_pattern": "TestSupportPhase0_",
      "fixture_policy": { "postgres": "migration_scratch" },
      "fixture_budget": {
        "postgres": {
          "max_migration_scratch": 1
        }
      }
    }
  ],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/app",
      "file": "internal/app/bootstrap_phase0_test.go",
      "symbol": "TestPhase0_BootstrapManifestValidation_U_0_07",
      "execution_dependency": "backend_unit",
      "evidence_layer": "fixture_policy_validation"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$missing_migration_reason_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase5 >/dev/null 2>&1; then
  fail "phase manifest validation must reject support migration_scratch without migration_scratch_reason"
fi
