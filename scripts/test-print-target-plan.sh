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
const rawPgtest = rows.find((row) => row.target === "backend-integration" && row.execution_family === "backend-integration-testutil");
if (!rawPgtest || rawPgtest.fixture_policy?.postgres !== "template_clone" || !Number.isInteger(rawPgtest.fixture_budget?.postgres?.max_template_clones)) {
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
const heavyIntegrationAggregates = plan.aggregates.filter((aggregate) => aggregate.weight_ms > 18000);
if (
  heavyIntegrationAggregates.length === 0 ||
  heavyIntegrationAggregates.some((aggregate) => aggregate.shards.length < 2)
) {
  process.exit(1);
}
if (!plan.shards.every((shard) => Number.isInteger(shard.shard_target_ms) && shard.shard_target_ms > 0)) {
  process.exit(1);
}
const integrationMultiItemShards = plan.shards.filter((shard) => shard.shard_target_ms === 18000 && (shard.has_authoritative || shard.has_support) && shard.item_count > 1);
if (!integrationMultiItemShards.every((shard) => shard.weight_ms <= 18000 && shard.shard_target_ms === 18000)) {
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
const rawTestutilItems = plan.shards
  .filter((shard) => shard.aggregate_name === "backend-integration-testutil")
  .flatMap((shard) => shard.items)
  .filter((item) => item.kind === "raw");
const rawPackages = rawTestutilItems.flatMap((item) => item.packages).sort();
const expectedRawPackages = [
  "./internal/testutil/httptestx",
  "./internal/testutil/pgtest",
  "./internal/testutil/s3test",
  "./internal/testutil/testcontainersx",
  "./internal/testutil/wstest",
];
if (
  rawTestutilItems.length !== expectedRawPackages.length ||
  rawPackages.join("\n") !== expectedRawPackages.join("\n") ||
  !rawTestutilItems.every((item) => item.baseline_key.includes("backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/"))
) {
  process.exit(1);
}
const shared = plan.shards.filter((shard) => shard.shared_across_targets);
if (shared.length !== 0) {
  process.exit(1);
}
if (!plan.shards.every((shard) => !(shard.has_authoritative && shard.has_support))) {
  process.exit(1);
}
EOF
then
  fail "backend-integration go shard plan must be weighted, policy-bearing, split heavy aggregates, and keep authoritative/support shards separate"
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
assert_contains "$backend_store_output" "execution_families=" "backend-store compact execution family count"

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
assert_contains "$make_output" "Cartulary target guidance: backend-store" "make explain-target"
assert_contains "$make_output" "phase_coverage:" "make explain-target summary"
if [[ -d "$results_dir" ]] && [[ -n "$(find "$results_dir" -mindepth 1 -print -quit)" ]]; then
  fail "make explain-target must not create test report artifacts"
fi

make_rows_output="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=rows)"
assert_contains "$make_rows_output" "phase1 unit authoritative" "make explain-target row mode"

make_target_plan_json="$tmp_dir/make-target-plan.json"
"$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" target-plan-json >"$make_target_plan_json"
"$NODE_HELPER" -e 'const plan = JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8")); if (!Array.isArray(plan) || plan.length === 0) process.exit(1);' "$make_target_plan_json"

phase_root="$tmp_dir/phase-root"
mkdir -p "$phase_root/tools"
cp "$ROOT_DIR"/tools/phase*_test_map.json "$phase_root/tools/"
cat >"$phase_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["U-99-01"],
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
      "execution_family": "backend-unit-auth",
      "execution_label": "Backend unit auth",
      "symbol": "TestSupportPhase5_Discovered"
    }
  ]
}
JSON

discovered_phases="$(CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" list-phases)"
assert_contains "$discovered_phases" "phase99" "phase manifest discovery includes phase99"

phase_map_discovery_root="$tmp_dir/phase-map-discovery-root"
mkdir -p "$phase_map_discovery_root/tools"
cat >"$phase_map_discovery_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_layer": "store_domain",
      "claim": "phase-map discovery validates future phase manifests",
      "out_of_scope": "phase-map discovery validates future phase manifests"
    }
  ]
}
JSON
phase5_check_maps_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_map_discovery_root" \
  NODE_BIN="$NODE_HELPER" \
    "$ROOT_DIR/scripts/check-phase-maps.sh"
)"
assert_contains "$phase5_check_maps_output" "phase5 traceability map verified" "check-phase-maps discovers phase5"

phase99_plan="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
    "$NODE_HELPER" "$PLAN_SCRIPT" --json
)"
assert_contains "$phase99_plan" '"manifest_phase": "phase99"' "target-plan support rows include discovered phase"

phase99_shared_command="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
  NODE_BIN="$NODE_HELPER" \
    "$NODE_HELPER" "$ROOT_DIR/scripts/cartulary-runner.mjs" go-target inspect-aggregate-command backend-unit backend-unit-auth
)"
assert_contains "$phase99_shared_command" "TestSupportPhase5_Discovered" "run-go-target support selection includes discovered phase"

invalid_phase_root="$tmp_dir/invalid-phase-root"
mkdir -p "$invalid_phase_root/tools"
cat >"$invalid_phase_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["U-99-01"],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase5_Invalid_U_5_01",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_layer": "store_domain",
      "fixture_policy": { "postgres": "invalid" },
      "claim": "invalid fixture policy smoke",
      "out_of_scope": "invalid fixture policy smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$invalid_phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 >/dev/null 2>&1; then
  fail "phase manifest validation must reject unknown postgres fixture policies"
fi

missing_policy_root="$tmp_dir/missing-policy-root"
mkdir -p "$missing_policy_root/tools"
cat >"$missing_policy_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["U-99-01"],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_layer": "store_domain",
      "claim": "missing fixture policy smoke",
      "out_of_scope": "missing fixture policy smoke"
    }
  ]
}
JSON

if ! CARTULARY_PHASE_MANIFEST_ROOT="$missing_policy_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 >/dev/null 2>&1; then
  fail "phase manifest validation must allow missing service-backed postgres fixture policies when defaults apply"
fi

missing_claim_root="$tmp_dir/missing-claim-root"
mkdir -p "$missing_claim_root/tools"
cat >"$missing_claim_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["U-99-01"],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_layer": "store_domain",
      "out_of_scope": "missing claim smoke"
    }
  ]
}
JSON

set +e
missing_claim_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$missing_claim_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 2>&1
)"
missing_claim_status=$?
set -e
if [[ "$missing_claim_status" -eq 0 ]]; then
  fail "phase manifest validation must reject authoritative entries without claim"
fi
assert_contains "$missing_claim_output" "must declare a non-empty claim" "missing claim validation output"

missing_budget_root="$tmp_dir/missing-budget-root"
mkdir -p "$missing_budget_root/tools"
cat >"$missing_budget_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["I-99-01"],
  "integration": [
    {
      "id": "I-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_integration_test.go",
      "symbol": "TestPhase1_LoginSessionLifecycle_I_1_01",
      "execution_dependency": "backend_integration",
      "execution_family": "backend-integration-auth",
      "execution_label": "Backend integration auth",
      "evidence_layer": "integration",
      "fixture_policy": { "postgres": "package_reset" },
      "claim": "missing fixture budget smoke",
      "out_of_scope": "missing fixture budget smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$missing_budget_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 >/dev/null 2>&1; then
  fail "phase manifest validation must reject package_reset without postgres fixture budgets"
fi

invalid_budget_root="$tmp_dir/invalid-budget-root"
mkdir -p "$invalid_budget_root/tools"
cat >"$invalid_budget_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["I-99-01"],
  "integration": [
    {
      "id": "I-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_integration_test.go",
      "symbol": "TestPhase1_LoginSessionLifecycle_I_1_01",
      "execution_dependency": "backend_integration",
      "execution_family": "backend-integration-auth",
      "execution_label": "Backend integration auth",
      "evidence_layer": "integration",
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

if CARTULARY_PHASE_MANIFEST_ROOT="$invalid_budget_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 >/dev/null 2>&1; then
  fail "phase manifest validation must reject invalid postgres fixture budgets"
fi

missing_migration_reason_root="$tmp_dir/missing-migration-reason-root"
mkdir -p "$missing_migration_reason_root/tools"
cat >"$missing_migration_reason_root/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["U-99-01"],
  "support_go_targets": [
    {
      "target": "backend_integration_support",
      "section": "integration",
      "package": "./internal/platform/objectstore",
      "file": "internal/platform/objectstore/objectstore_phase0_support_test.go",
      "symbol": "TestSupportPhase0_ManagedServiceObjectStoreBinding",
      "selection_pattern": "TestSupportPhase0_",
      "execution_family": "backend-integration-platform",
      "execution_label": "Backend integration platform",
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
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/app",
      "file": "internal/app/bootstrap_phase0_test.go",
      "symbol": "TestPhase0_BootstrapManifestValidation_U_0_07",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_layer": "fixture_policy_validation",
      "claim": "synthetic migration scratch validation smoke",
      "out_of_scope": "synthetic migration scratch validation smoke"
    }
  ]
}
JSON

if CARTULARY_PHASE_MANIFEST_ROOT="$missing_migration_reason_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 >/dev/null 2>&1; then
  fail "phase manifest validation must reject support migration_scratch without migration_scratch_reason"
fi

ledger_root="$tmp_dir/ledger-root"
mkdir -p "$ledger_root/tools" "$ledger_root/internal/modules/auth" "$ledger_root/cmd/server"
cat >"$ledger_root/internal/modules/auth/phase99_test.go" <<'EOF'
package auth

func TestPhase99_Ledger_U_99_01() {}
EOF
cat >"$ledger_root/tools/phase99_test_map.json" <<'JSON'
{
  "note": "Synthetic ledger phase.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "scope": "synthetic future phase ledger smoke.",
    "normative_owners": "Synthetic owner.",
    "notes": ["Synthetic note."],
    "authoritative_execution": [
      "`backend-unit` selects authoritative `U-99-*` rows through manifest discovery."
    ],
    "support_execution_extras": [],
    "sections": {
      "unit": "Unit"
    },
    "shared_harness": [
      "| Harness | Phase 99 evidence |",
      "| --- | --- |",
      "| Synthetic harness | Synthetic future phase ledger smoke. |"
    ],
    "support_only": [
      "Synthetic support-only evidence."
    ]
  },
  "expected_ids": ["U-99-01"],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase99_test.go",
      "symbol": "TestPhase99_Ledger_U_99_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-auth",
      "execution_label": "Backend unit auth",
      "evidence_layer": "ledger_smoke",
      "claim": "synthetic future phases render ledger metadata without renderer code changes",
      "out_of_scope": "synthetic future phases render ledger metadata without renderer code changes"
    }
  ]
}
JSON

ledger_render_output="$(
  cd "$ledger_root"
  CARTULARY_PHASE_MANIFEST_ROOT="$ledger_root" "$NODE_HELPER" "$ROOT_DIR/scripts/render-phase-ledgers.mjs"
)"
assert_contains "$ledger_render_output" "rendered docs/testing/phase99_coverage_ledger.md" "future phase ledger render"
ledger_drift_output="$(
  cd "$ledger_root"
  CARTULARY_PHASE_MANIFEST_ROOT="$ledger_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-ledger-drift.mjs"
)"
assert_contains "$ledger_drift_output" "phase coverage ledgers verified" "future phase ledger drift check"
