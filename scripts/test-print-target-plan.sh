#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
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

write_phase_registry() {
  local root="$1"
  shift

  mkdir -p "$root/tools" "$root/internal" "$root/cmd/server"
  {
    printf '{\n  "schema_id": "cartulary.phase_registry.v1",\n  "phases": [\n'
    local first=1
    local phase
    for phase in "$@"; do
      local phase_number="${phase#phase}"
      if [[ "$first" -eq 0 ]]; then
        printf ',\n'
      fi
      first=0
      printf '    {\n'
      printf '      "phase": "%s",\n' "$phase"
      printf '      "order": %s,\n' "$phase_number"
      printf '      "status": "active",\n'
      printf '      "label": "Phase %s",\n' "$phase_number"
      printf '      "manifest_path": "tools/%s_test_map.json",\n' "$phase"
      printf '      "ledger_path": "docs/testing/%s_coverage_ledger.md",\n' "$phase"
      printf '      "scope": "synthetic %s scope.",\n' "$phase"
      printf '      "normative_owners": "Synthetic owner."\n'
      printf '    }'
    done
    printf '\n  ]\n}\n'
  } >"$root/tools/phase_registry.json"
}

write_phase_ledger_stub() {
  local root="$1"
  local phase="$2"

  mkdir -p "$root/docs/testing"
  printf '# %s\n' "$phase" >"$root/docs/testing/${phase}_coverage_ledger.md"
}

write_go_source_symbol() {
  local root="$1"
  local file="$2"
  local package_name="$3"
  local symbol="$4"

  mkdir -p "$(dirname "$root/$file")"
  {
    printf 'package %s\n\n' "$package_name"
    printf 'func %s() {}\n' "$symbol"
  } >"$root/$file"
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
const validPolicies = new Set(["template_clone", "package_reset", "migration_scratch", "transaction", "group_clone"]);
if (storeRows.length === 0 || !storeRows.every((row) => validPolicies.has(row.fixture_policy?.postgres))) {
  process.exit(1);
}
const storeTemplateCloneRows = storeRows.filter((row) => row.fixture_policy?.postgres === "template_clone");
if (!storeTemplateCloneRows.every((row) => Number.isInteger(row.fixture_budget?.postgres?.max_template_clones))) {
  process.exit(1);
}
const rawPgtest = rows.find((row) => row.target === "backend-integration" && row.execution_family === "backend-integration-testutil");
if (!rawPgtest || rawPgtest.fixture_policy?.postgres !== "template_clone" || !Number.isInteger(rawPgtest.fixture_budget?.postgres?.max_template_clones)) {
  process.exit(1);
}
const serviceBackedGoRows = rows.filter((row) => row.service_backed && row.runner_family === "go_test");
if (serviceBackedGoRows.length === 0 || !serviceBackedGoRows.every((row) => validPolicies.has(row.fixture_policy?.postgres))) {
  process.exit(1);
}
const packageResetRows = serviceBackedGoRows.filter((row) => row.fixture_policy?.postgres === "package_reset" && row.coverage !== "raw");
if (!packageResetRows.every((row) => Number.isInteger(row.fixture_budget?.postgres?.max_package_resets))) {
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
if (!packageReset.every((item) => Number.isInteger(item.postgres_fixture_budget?.max_package_resets))) {
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
const validPolicies = new Set(["template_clone", "transaction"]);
if (items.length === 0 || !items.every((item) => item.kind === "authoritative" && validPolicies.has(item.postgres_fixture_policy))) {
  process.exit(1);
}
const transactionItems = items.filter((item) => item.postgres_fixture_policy === "transaction");
if (!transactionItems.every((item) => Number.isInteger(item.postgres_fixture_budget?.max_transactions))) {
  process.exit(1);
}
const templateCloneItems = items.filter((item) => item.postgres_fixture_policy === "template_clone");
if (!templateCloneItems.every((item) => Number.isInteger(item.postgres_fixture_budget?.max_template_clones))) {
  process.exit(1);
}
EOF
then
  fail "backend-store go shard plan must expose authoritative fixture planning"
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
write_phase_registry "$phase_root" phase0 phase1 phase2 phase3 phase4 phase99
cat >"$phase_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic target-plan support discovery fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic target-plan support discovery fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-05"],
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
      "evidence_class": "implementation_support",
      "layer": "backend_unit",
      "default_check_required": false,
      "default_check_kind": "explicit_only",
      "default_check_reason_code": "implementation_support_explicit_only",
      "primary_evidence_owner": "backend_unit::internal/modules/auth/phase1_support_test.go::TestSupportPhase5_",
      "duplicate_of": null,
      "evidence_delta": "support evidence is explicit-only",
      "warm_local_cost_class": "low",
      "symbol": "TestSupportPhase5_Discovered"
    }
  ]
}
JSON

discovered_phases="$(CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" list-phases)"
assert_contains "$discovered_phases" "phase99" "phase registry discovery includes phase99"

phase_map_discovery_root="$tmp_dir/phase-map-discovery-root"
mkdir -p "$phase_map_discovery_root/tools"
write_phase_registry "$phase_map_discovery_root" phase99
write_phase_ledger_stub "$phase_map_discovery_root" phase99
write_go_source_symbol "$phase_map_discovery_root" "internal/modules/auth/phase1_store_test.go" "auth" "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01"
cat >"$phase_map_discovery_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic phase-map discovery fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic phase-map discovery fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-99-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_class": "product_conformance",
      "layer": "backend_store",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_store::internal/modules/auth/phase1_store_test.go::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic phase-map discovery fixture coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "store_domain",
      "claim_status": "implemented",
      "claim": "phase-map discovery validates future phase manifests",
      "out_of_scope": "phase-map discovery validates future phase manifests"
    }
  ],
  "integration": [],
  "e2e": []
}
JSON
phase99_check_maps_output="$(
  cd "$phase_map_discovery_root"
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_map_discovery_root" \
  NODE_BIN="$NODE_HELPER" \
    "$ROOT_DIR/scripts/check-phase-maps.sh"
)"
assert_contains "$phase99_check_maps_output" "phase99 traceability map verified" "check-phase-maps validates registry phase99"

registry_order_root="$tmp_dir/registry-order-root"
mkdir -p "$registry_order_root/tools"
cat >"$registry_order_root/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase12",
      "order": 2,
      "status": "active",
      "label": "Phase 12",
      "manifest_path": "tools/phase12_test_map.json",
      "ledger_path": "docs/testing/phase12_coverage_ledger.md",
      "scope": "synthetic phase12 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase2",
      "order": 1,
      "status": "active",
      "label": "Phase 2",
      "manifest_path": "tools/phase2_test_map.json",
      "ledger_path": "docs/testing/phase2_coverage_ledger.md",
      "scope": "synthetic phase2 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase99",
      "order": 99,
      "status": "planned",
      "label": "Phase 99",
      "manifest_path": "tools/phase99_test_map.json",
      "ledger_path": "docs/testing/phase99_coverage_ledger.md",
      "scope": "synthetic planned scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON
ordered_phases="$(CARTULARY_PHASE_MANIFEST_ROOT="$registry_order_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" list-phases)"
if [[ "$ordered_phases" != $'phase2\nphase12' ]]; then
  fail "phase registry order/status: expected active order phase2 then phase12, got [$ordered_phases]"
fi
planned_explain="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$registry_order_root" "$NODE_HELPER" "$ROOT_DIR/scripts/print-explain-phase.mjs" --phase phase99
)"
assert_contains "$planned_explain" "Cartulary phase guidance: phase99" "planned phase explain"
set +e
planned_slice_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$registry_order_root" "$NODE_HELPER" "$ROOT_DIR/scripts/run-phase-slice.mjs" --phase phase99 --mode phase --json 2>&1
)"
planned_slice_status=$?
set -e
if [[ "$planned_slice_status" -eq 0 ]]; then
  fail "planned phase slice must fail"
fi
assert_contains "$planned_slice_output" "phase phase99 is planned and is not executable" "planned phase slice output"

registry_bad_schema_root="$tmp_dir/registry-bad-schema-root"
mkdir -p "$registry_bad_schema_root/tools"
cat >"$registry_bad_schema_root/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v0",
  "phases": []
}
JSON
set +e
bad_schema_output="$(CARTULARY_PHASE_MANIFEST_ROOT="$registry_bad_schema_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-registry.mjs" validate 2>&1)"
bad_schema_status=$?
set -e
if [[ "$bad_schema_status" -eq 0 ]]; then
  fail "phase registry validation must reject wrong schema"
fi
assert_contains "$bad_schema_output" "must declare schema_id cartulary.phase_registry.v1" "bad registry schema output"

registry_bad_path_root="$tmp_dir/registry-bad-path-root"
mkdir -p "$registry_bad_path_root/tools"
cat >"$registry_bad_path_root/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase1",
      "order": 1,
      "status": "active",
      "label": "Phase 1",
      "manifest_path": "../phase1_test_map.json",
      "ledger_path": "docs/testing/phase1_coverage_ledger.md",
      "scope": "synthetic phase1 scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON
set +e
bad_path_output="$(CARTULARY_PHASE_MANIFEST_ROOT="$registry_bad_path_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-registry.mjs" validate 2>&1)"
bad_path_status=$?
set -e
if [[ "$bad_path_status" -eq 0 ]]; then
  fail "phase registry validation must reject path traversal"
fi
assert_contains "$bad_path_output" "manifest_path must not escape the repository root" "bad registry path output"

registry_orphan_root="$tmp_dir/registry-orphan-root"
mkdir -p "$registry_orphan_root/tools" "$registry_orphan_root/docs/testing"
write_phase_registry "$registry_orphan_root" phase1
write_phase_ledger_stub "$registry_orphan_root" phase1
cat >"$registry_orphan_root/tools/phase1_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase1",
  "note": "Synthetic registry orphan fixture.",
  "ledger": {
    "title": "Phase 1 Coverage Ledger",
    "notes": "Synthetic registry orphan fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase1",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
cat >"$registry_orphan_root/tools/phase2_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase2",
  "note": "Synthetic registry orphan fixture.",
  "ledger": {
    "title": "Phase 2 Coverage Ledger",
    "notes": "Synthetic registry orphan fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase2",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-2-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
set +e
orphan_output="$(CARTULARY_PHASE_MANIFEST_ROOT="$registry_orphan_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-registry.mjs" validate 2>&1)"
orphan_status=$?
set -e
if [[ "$orphan_status" -eq 0 ]]; then
  fail "phase registry validation must reject unregistered phase maps"
fi
assert_contains "$orphan_output" "unregistered phase test map: tools/phase2_test_map.json" "orphan phase map output"

registry_missing_active_root="$tmp_dir/registry-missing-active-root"
mkdir -p "$registry_missing_active_root/tools"
write_phase_registry "$registry_missing_active_root" phase1
set +e
missing_active_output="$(CARTULARY_PHASE_MANIFEST_ROOT="$registry_missing_active_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-registry.mjs" validate 2>&1)"
missing_active_status=$?
set -e
if [[ "$missing_active_status" -eq 0 ]]; then
  fail "phase registry validation must reject missing active artifacts"
fi
assert_contains "$missing_active_output" "active phase1 manifest missing" "missing active manifest output"

registry_retired_root="$tmp_dir/registry-retired-root"
mkdir -p "$registry_retired_root/tools"
cat >"$registry_retired_root/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase7",
      "order": 7,
      "status": "retired",
      "label": "Phase 7",
      "manifest_path": "tools/phase7_test_map.json",
      "ledger_path": "docs/testing/phase7_coverage_ledger.md",
      "scope": "synthetic retired scope.",
      "normative_owners": "Synthetic owner.",
      "retired_reason": "synthetic retirement smoke"
    }
  ]
}
JSON
set +e
retired_output="$(CARTULARY_PHASE_MANIFEST_ROOT="$registry_retired_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-registry.mjs" validate 2>&1)"
retired_status=$?
set -e
if [[ "$retired_status" -eq 0 ]]; then
  fail "phase registry validation must reject retired phases without retained artifacts"
fi
assert_contains "$retired_output" "retained_artifacts must be a non-empty array for retired phases" "retired registry output"

assert_phase_identity_rejected() {
  local root="$1"
  local phase="$2"
  local expected="$3"
  local label="$4"
  local output
  local status

  set +e
  output="$(
    CARTULARY_PHASE_MANIFEST_ROOT="$root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" "$phase" 2>&1
  )"
  status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected phase manifest identity validation failure"
  fi
  assert_contains "$output" "$expected" "$label"
}

missing_schema_identity_root="$tmp_dir/identity-missing-schema-root"
mkdir -p "$missing_schema_identity_root/tools"
write_phase_registry "$missing_schema_identity_root" phase99
cat >"$missing_schema_identity_root/tools/phase99_test_map.json" <<'JSON'
{
  "phase": "phase99",
  "note": "Synthetic identity fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic identity fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-05"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
assert_phase_identity_rejected "$missing_schema_identity_root" "phase99" "must declare schema_id cartulary.phase_test_map.v2" "missing phase-map schema_id"

wrong_schema_identity_root="$tmp_dir/identity-wrong-schema-root"
mkdir -p "$wrong_schema_identity_root/tools"
write_phase_registry "$wrong_schema_identity_root" phase99
cat >"$wrong_schema_identity_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v0",
  "phase": "phase99",
  "note": "Synthetic identity fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic identity fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-05"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
assert_phase_identity_rejected "$wrong_schema_identity_root" "phase99" "must declare schema_id cartulary.phase_test_map.v2" "wrong phase-map schema_id"

missing_phase_identity_root="$tmp_dir/identity-missing-phase-root"
mkdir -p "$missing_phase_identity_root/tools"
write_phase_registry "$missing_phase_identity_root" phase99
cat >"$missing_phase_identity_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "note": "Synthetic identity fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic identity fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-05"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
assert_phase_identity_rejected "$missing_phase_identity_root" "phase99" ".phase must be a non-empty string" "missing phase-map phase"

unknown_key_identity_root="$tmp_dir/identity-unknown-key-root"
mkdir -p "$unknown_key_identity_root/tools"
write_phase_registry "$unknown_key_identity_root" phase99
cat >"$unknown_key_identity_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic shape validation fixture.",
  "legacy_manifest_key": true,
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic shape validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-05"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
assert_phase_identity_rejected "$unknown_key_identity_root" "phase99" "unknown key legacy_manifest_key" "semantic phase-map path rejects unknown shape key"

stale_title_identity_root="$tmp_dir/identity-stale-title-root"
mkdir -p "$stale_title_identity_root/tools"
write_phase_registry "$stale_title_identity_root" phase99
cat >"$stale_title_identity_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic stale authoritative title fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic stale authoritative title fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["E-99-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": [
    {
      "id": "E-99-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/phase99.spec.ts",
      "title": "Phase 99 stale browser title",
      "execution_dependency": "browser_functional",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_reason": "Default browser projection proves cross-stack workflow behavior that lower layers cannot fully exercise.",
      "default_check_kind": "default_local_cross_stack_conformance",
      "default_check_reason_code": "lower_layer_gap",
      "primary_evidence_owner": "browser_functional::apps/web/e2e/phase99.spec.ts::Phase 99 stale browser title",
      "duplicate_of": null,
      "evidence_delta": "Synthetic stale authoritative title fixture coverage.",
      "warm_local_cost_class": "browser",
      "evidence_layer": "browser",
      "claim_status": "implemented",
      "claim": "future browser evidence",
      "out_of_scope": "none"
    }
  ]
}
JSON
assert_phase_identity_rejected "$stale_title_identity_root" "phase99" "authoritative evidence for E-99-01 must include E-99-01 or E_99_01" "semantic phase-map path rejects stale authoritative title"

mismatched_phase_identity_root="$tmp_dir/identity-mismatched-phase-root"
mkdir -p "$mismatched_phase_identity_root/tools"
write_phase_registry "$mismatched_phase_identity_root" phase99
cat >"$mismatched_phase_identity_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase98",
  "note": "Synthetic identity fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic identity fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-05"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
assert_phase_identity_rejected "$mismatched_phase_identity_root" "phase99" "declares phase phase98 but filename declares phase99" "mismatched phase-map phase"

leading_zero_identity_root="$tmp_dir/identity-leading-zero-root"
mkdir -p "$leading_zero_identity_root/tools"
cat >"$leading_zero_identity_root/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase01",
      "order": 1,
      "status": "active",
      "label": "Phase 01",
      "manifest_path": "tools/phase01_test_map.json",
      "ledger_path": "docs/testing/phase01_coverage_ledger.md",
      "scope": "synthetic phase01 scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON
cat >"$leading_zero_identity_root/tools/phase01_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase01",
  "note": "Synthetic identity fixture.",
  "ledger": {
    "title": "Phase 01 Coverage Ledger",
    "notes": "Synthetic identity fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase01",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-1-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
assert_phase_identity_rejected "$leading_zero_identity_root" "phase01" "invalid phase name phase01" "leading-zero phase-map phase"

phase99_plan="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
    "$NODE_HELPER" "$PLAN_SCRIPT" --json
)"
assert_contains "$phase99_plan" '"manifest_phase": "phase99"' "target-plan support rows include registry phase"

phase99_shared_command="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
  NODE_BIN="$NODE_HELPER" \
    "$NODE_HELPER" "$ROOT_DIR/scripts/cartulary-runner.mjs" go-target inspect-aggregate-command backend-unit backend-unit-auth
)"
assert_contains "$phase99_shared_command" "TestSupportPhase5_Discovered" "run-go-target support selection includes registry phase"

invalid_phase_root="$tmp_dir/invalid-phase-root"
mkdir -p "$invalid_phase_root/tools"
write_phase_registry "$invalid_phase_root" phase99
write_go_source_symbol "$invalid_phase_root" "internal/modules/auth/phase1_store_test.go" "auth" "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01"
cat >"$invalid_phase_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic fixture policy validation fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic fixture policy validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-99-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_class": "product_conformance",
      "layer": "backend_store",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_store::internal/modules/auth/phase1_store_test.go::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic invalid fixture policy coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "store_domain",
      "claim_status": "implemented",
      "fixture_policy": { "postgres": "invalid" },
      "claim": "invalid fixture policy smoke",
      "out_of_scope": "invalid fixture policy smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
JSON

if (cd "$invalid_phase_root" && CARTULARY_PHASE_MANIFEST_ROOT="$invalid_phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99) >/dev/null 2>&1; then
  fail "phase manifest validation must reject unknown postgres fixture policies"
fi

missing_policy_root="$tmp_dir/missing-policy-root"
mkdir -p "$missing_policy_root/tools"
write_phase_registry "$missing_policy_root" phase99
write_go_source_symbol "$missing_policy_root" "internal/modules/auth/phase1_store_test.go" "auth" "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01"
cat >"$missing_policy_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic fixture policy validation fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic fixture policy validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-99-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_class": "product_conformance",
      "layer": "backend_store",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_store::internal/modules/auth/phase1_store_test.go::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic missing fixture policy coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "store_domain",
      "claim_status": "implemented",
      "claim": "missing fixture policy smoke",
      "out_of_scope": "missing fixture policy smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
JSON

if ! (cd "$missing_policy_root" && CARTULARY_PHASE_MANIFEST_ROOT="$missing_policy_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99) >/dev/null 2>&1; then
  fail "phase manifest validation must allow missing service-backed postgres fixture policies when defaults apply"
fi

missing_claim_root="$tmp_dir/missing-claim-root"
mkdir -p "$missing_claim_root/tools"
write_phase_registry "$missing_claim_root" phase99
write_go_source_symbol "$missing_claim_root" "internal/modules/auth/phase1_store_test.go" "auth" "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01"
cat >"$missing_claim_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic fixture policy validation fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic fixture policy validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-99-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_class": "product_conformance",
      "layer": "backend_store",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_store::internal/modules/auth/phase1_store_test.go::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic missing claim coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "store_domain",
      "claim_status": "implemented",
      "out_of_scope": "missing claim smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
JSON

set +e
missing_claim_output="$(
  cd "$missing_claim_root"
  CARTULARY_PHASE_MANIFEST_ROOT="$missing_claim_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 2>&1
)"
missing_claim_status=$?
set -e
if [[ "$missing_claim_status" -eq 0 ]]; then
  fail "phase manifest validation must reject authoritative entries without claim"
fi
assert_contains "$missing_claim_output" "must declare a non-empty claim" "missing claim validation output"

blocked_profile_claim_root="$tmp_dir/blocked-profile-claim-root"
mkdir -p "$blocked_profile_claim_root/tools"
write_phase_registry "$blocked_profile_claim_root" phase99
write_go_source_symbol "$blocked_profile_claim_root" "internal/modules/auth/phase1_store_test.go" "auth" "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01"
cat >"$blocked_profile_claim_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic profile claim validation fixture.",
  "profile_claims": [
    {
      "profile_id": "synthetic_extension",
      "claimed": true,
      "claim_ac_id": "AC-999",
      "required_ac_ids": ["AC-999"],
      "direct_evidence_ids": ["U-99-01"],
      "aggregate_ac_ids": ["AC-999"]
    }
  ],
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic profile claim validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-99-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_class": "product_conformance",
      "layer": "backend_store",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_store::internal/modules/auth/phase1_store_test.go::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic blocked profile evidence coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "store_domain",
      "claim_status": "blocked",
      "claim": "blocked profile evidence smoke",
      "out_of_scope": "blocked profile evidence smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
JSON

set +e
blocked_profile_claim_output="$(
  cd "$blocked_profile_claim_root"
  CARTULARY_PHASE_MANIFEST_ROOT="$blocked_profile_claim_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99 2>&1
)"
blocked_profile_claim_status=$?
set -e
if [[ "$blocked_profile_claim_status" -eq 0 ]]; then
  fail "phase manifest validation must reject claimed profiles with blocked direct evidence"
fi
assert_contains "$blocked_profile_claim_output" "direct_evidence_id U-99-01 must have claim_status=implemented" "blocked profile claim validation output"

missing_budget_root="$tmp_dir/missing-budget-root"
mkdir -p "$missing_budget_root/tools"
write_phase_registry "$missing_budget_root" phase99
write_go_source_symbol "$missing_budget_root" "internal/modules/auth/phase1_integration_test.go" "auth" "TestPhase1_LoginSessionLifecycle_I_99_01"
cat >"$missing_budget_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic fixture budget validation fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic fixture budget validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["I-99-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [
    {
      "id": "I-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_integration_test.go",
      "symbol": "TestPhase1_LoginSessionLifecycle_I_99_01",
      "execution_dependency": "backend_integration",
      "execution_family": "backend-integration-auth",
      "execution_label": "Backend integration auth",
      "evidence_class": "product_conformance",
      "layer": "backend_integration",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_integration::internal/modules/auth/phase1_integration_test.go::TestPhase1_LoginSessionLifecycle_I_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic missing fixture budget coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "integration",
      "claim_status": "implemented",
      "fixture_policy": { "postgres": "package_reset" },
      "claim": "missing fixture budget smoke",
      "out_of_scope": "missing fixture budget smoke"
    }
  ],
  "e2e": []
}
JSON

if (cd "$missing_budget_root" && CARTULARY_PHASE_MANIFEST_ROOT="$missing_budget_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99) >/dev/null 2>&1; then
  fail "phase manifest validation must reject package_reset without postgres fixture budgets"
fi

invalid_budget_root="$tmp_dir/invalid-budget-root"
mkdir -p "$invalid_budget_root/tools"
write_phase_registry "$invalid_budget_root" phase99
write_go_source_symbol "$invalid_budget_root" "internal/modules/auth/phase1_integration_test.go" "auth" "TestPhase1_LoginSessionLifecycle_I_99_01"
cat >"$invalid_budget_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic fixture budget validation fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic fixture budget validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["I-99-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [
    {
      "id": "I-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_integration_test.go",
      "symbol": "TestPhase1_LoginSessionLifecycle_I_99_01",
      "execution_dependency": "backend_integration",
      "execution_family": "backend-integration-auth",
      "execution_label": "Backend integration auth",
      "evidence_class": "product_conformance",
      "layer": "backend_integration",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_integration::internal/modules/auth/phase1_integration_test.go::TestPhase1_LoginSessionLifecycle_I_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic invalid fixture budget coverage.",
      "warm_local_cost_class": "service_backed",
      "evidence_layer": "integration",
      "claim_status": "implemented",
      "fixture_policy": { "postgres": "package_reset" },
      "fixture_budget": {
        "postgres": {
          "max_package_resets": 1,
          "dirty_tables": ["users", "users"]
        }
      },
      "claim": "invalid fixture budget smoke",
      "out_of_scope": "invalid fixture budget smoke"
    }
  ],
  "e2e": []
}
JSON

if (cd "$invalid_budget_root" && CARTULARY_PHASE_MANIFEST_ROOT="$invalid_budget_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99) >/dev/null 2>&1; then
  fail "phase manifest validation must reject invalid postgres fixture budgets"
fi

missing_migration_reason_root="$tmp_dir/missing-migration-reason-root"
mkdir -p "$missing_migration_reason_root/tools"
write_phase_registry "$missing_migration_reason_root" phase99
write_go_source_symbol "$missing_migration_reason_root" "internal/platform/bootstrap/bootstrap_phase0_test.go" "bootstrap" "TestPhase0_BootstrapManifestValidation_U_99_01"
cat >"$missing_migration_reason_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic migration scratch validation fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic migration scratch validation fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
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
      "evidence_class": "implementation_support",
      "layer": "backend_integration_support",
      "default_check_required": false,
      "default_check_kind": "explicit_only",
      "default_check_reason_code": "implementation_support_explicit_only",
      "primary_evidence_owner": "backend_integration_support::internal/platform/objectstore/objectstore_phase0_support_test.go::TestSupportPhase0_",
      "duplicate_of": null,
      "evidence_delta": "support evidence is explicit-only",
      "warm_local_cost_class": "service_backed",
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
      "package": "./internal/platform/bootstrap",
      "file": "internal/platform/bootstrap/bootstrap_phase0_test.go",
      "symbol": "TestPhase0_BootstrapManifestValidation_U_99_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_class": "product_conformance",
      "layer": "backend_unit",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_unit::internal/platform/bootstrap/bootstrap_phase0_test.go::TestPhase0_BootstrapManifestValidation_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic migration scratch validation coverage.",
      "warm_local_cost_class": "low",
      "evidence_layer": "fixture_policy_validation",
      "claim_status": "implemented",
      "claim": "synthetic migration scratch validation smoke",
      "out_of_scope": "synthetic migration scratch validation smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
JSON

if (cd "$missing_migration_reason_root" && CARTULARY_PHASE_MANIFEST_ROOT="$missing_migration_reason_root" "$NODE_HELPER" "$ROOT_DIR/scripts/check-phase-map.mjs" phase99) >/dev/null 2>&1; then
  fail "phase manifest validation must reject support migration_scratch without migration_scratch_reason"
fi

ledger_root="$tmp_dir/ledger-root"
mkdir -p "$ledger_root/tools" "$ledger_root/internal/modules/auth" "$ledger_root/cmd/server"
write_phase_registry "$ledger_root" phase99
cat >"$ledger_root/internal/modules/auth/phase99_test.go" <<'EOF'
package auth

func TestPhase99_Ledger_U_99_01() {}
EOF
cat >"$ledger_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase99",
  "note": "Synthetic ledger phase.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
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
  "support_go_targets": [],
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
      "evidence_class": "product_conformance",
      "layer": "backend_unit",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_unit::internal/modules/auth/phase99_test.go::TestPhase99_Ledger_U_99_01",
      "duplicate_of": null,
      "evidence_delta": "Synthetic future phase ledger coverage.",
      "warm_local_cost_class": "low",
      "evidence_layer": "ledger_smoke",
      "claim_status": "implemented",
      "claim": "synthetic future phases render ledger metadata without renderer code changes",
      "out_of_scope": "synthetic future phases render ledger metadata without renderer code changes"
    }
  ],
  "integration": [],
  "e2e": []
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
