#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_HELPER="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
PLAN_SCRIPT="$ROOT_DIR/tools/harness/diagnostics/target-plan-cli.mjs"
SHARD_PLAN_SCRIPT="$ROOT_DIR/tools/harness/backend/go-shard-plan-cli.mjs"
MODE="${1:-all}"
cleanup_paths=()
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

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

case "$MODE" in
  all | diagnostics | backend) ;;
  *)
    fail "usage: target-plan-smoke-suite.sh [all|diagnostics|backend]"
    ;;
esac

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

tmp_dir="$(cartulary_harness_mktemp_dir "target-plan-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")

if [[ "$MODE" == "all" || "$MODE" == "diagnostics" ]]; then
json_a="$tmp_dir/target-plan-a.json"
json_b="$tmp_dir/target-plan-b.json"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_a"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_b"

"$NODE_HELPER" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$json_a"
cmp -s "$json_a" "$json_b" || fail "target-plan JSON must be deterministic across invocations"

if ! "$NODE_HELPER" - "$json_a" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const [jsonPath] = process.argv.slice(2);
const rows = JSON.parse(fs.readFileSync(jsonPath, "utf8"));
const graphUnit = rows.find(
  (row) =>
    row.owner_id === "module.graphprojection" &&
    row.id === "module.graphprojection.engine.canonical_behavior" &&
    row.target === "backend-unit",
);
const graphStore = rows.find(
  (row) =>
    row.owner_id === "module.graphprojection" &&
    row.id === "module.graphprojection.storage.lifecycle" &&
    row.target === "backend-store",
);
if (
  !graphUnit ||
  !graphStore ||
  graphUnit.packages?.[0] !== "./internal/modules/graphprojection" ||
  graphStore.fixture_policy?.postgres !== "transaction"
) {
  process.exit(1);
}
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
const serviceBackedPostgresRows = serviceBackedGoRows.filter((row) => Number(row.resource_claims?.postgres ?? 0) > 0);
if (
  serviceBackedGoRows.length === 0 ||
  serviceBackedPostgresRows.length === 0 ||
  !serviceBackedPostgresRows.every((row) => validPolicies.has(row.fixture_policy?.postgres))
) {
  process.exit(1);
}
const packageResetRows = serviceBackedPostgresRows.filter((row) => row.fixture_policy?.postgres === "package_reset" && row.coverage !== "raw");
if (!packageResetRows.every((row) => Number.isInteger(row.fixture_budget?.postgres?.max_package_resets))) {
  process.exit(1);
}
const transactionRows = serviceBackedPostgresRows.filter((row) => row.fixture_policy?.postgres === "transaction");
if (!transactionRows.every((row) => Number.isInteger(row.fixture_budget?.postgres?.max_transactions))) {
  process.exit(1);
}
EOF
then
  fail "target-plan JSON must expose semantic catalog identities and postgres fixture policies"
fi
fi

if [[ "$MODE" == "all" || "$MODE" == "backend" ]]; then
"$NODE_HELPER" "$ROOT_DIR/tools/harness/backend/tests/test-service-go-batching.mjs"
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
const integrationMultiItemShards = plan.shards.filter(
  (shard) => (shard.has_authoritative || shard.has_support) && shard.item_count > 1,
);
if (
  integrationMultiItemShards.length === 0 ||
  !integrationMultiItemShards.every(
    (shard) =>
      shard.work_weight_ms <= 12000 &&
      shard.shard_target_ms === 12000 &&
      shard.item_count <= 8 &&
      new Set(shard.items.map((item) => item.target)).size === 1 &&
      new Set(shard.items.map((item) => JSON.stringify(item.package_import_paths))).size === 1 &&
      new Set(shard.items.map((item) => JSON.stringify(item.runtime_binaries))).size === 1 &&
      new Set(shard.items.map((item) => item.postgres_fixture_policy)).size === 1 &&
      new Set(shard.items.map((item) => JSON.stringify(item.postgres_fixture_budget))).size === 1,
  )
) {
  process.exit(1);
}
const authoritative = plan.shards.flatMap((shard) => shard.items).filter((item) => item.kind === "authoritative");
const postgresAuthoritative = authoritative.filter((item) => item.postgres_fixture_policy !== "");
const validPolicies = new Set(["template_clone", "package_reset", "migration_scratch", "transaction", "group_clone"]);
if (
  authoritative.length === 0 ||
  postgresAuthoritative.length === 0 ||
  !postgresAuthoritative.every((item) => validPolicies.has(item.postgres_fixture_policy))
) {
  process.exit(1);
}
const packageReset = postgresAuthoritative.filter((item) => item.postgres_fixture_policy === "package_reset");
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

backend_process_shard_json="$tmp_dir/go-shard-plan-backend-process.json"
"$NODE_HELPER" "$SHARD_PLAN_SCRIPT" --json --target backend-process >"$backend_process_shard_json"
if ! "$NODE_HELPER" - "$backend_process_shard_json" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const [jsonPath] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(jsonPath, "utf8"));
const executionFamily = "module.recovery.process";
const recoveryShards = plan.shards
  .filter((shard) => shard.aggregate_name === executionFamily)
  .sort((left, right) => left.name.localeCompare(right.name));
const recoveryItems = recoveryShards.flatMap((shard) => shard.items ?? []);
const expectedRows = [
  "module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9",
  "module.recovery.process.the_standalone_operator_initializes_the_configur_03834eb65a",
];
const actualRows = [...new Set(recoveryItems.map((item) => item.id))].sort();
if (
  JSON.stringify(actualRows) !== JSON.stringify(expectedRows) ||
  recoveryItems.length !== 6 ||
  recoveryItems.some(
    (item) => item.primary_evidence_owner !== item.id || item.scenario_id !== ""
  ) ||
  recoveryShards.length >= recoveryItems.length ||
  recoveryShards.some(
    (shard) =>
      shard.item_count < 1 ||
      shard.item_count > 8 ||
      shard.work_weight_ms > 12000 ||
      shard.shard_target_ms !== 12000,
  )
) {
  process.exit(1);
}
EOF
then
  fail "backend-process go shard plan must batch semantic recovery rows deterministically with exact selector evidence"
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
const validPolicies = new Set(["template_clone", "package_reset", "migration_scratch", "transaction", "group_clone"]);
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
fi

if [[ "$MODE" == "all" || "$MODE" == "diagnostics" ]]; then
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
assert_contains "$backend_store_detail" "module.graphprojection.storage.lifecycle:" "backend-store detailed target plan"
assert_contains "$backend_store_detail" "module.auth.store.one_route_matrix_proves_logout_password_change_t_728b4fb4df:" "backend-store detailed target plan"
assert_contains "$backend_store_detail" "packages:" "backend-store detail packages"

results_dir="$tmp_dir/results"
make_output="$(
  env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET \
    CARTULARY_TEST_RESULTS_DIR="$results_dir" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store
)"
assert_contains "$make_output" "Cartulary target guidance: backend-store" "make explain-target"
assert_contains "$make_output" "step_coverage:" "make explain-target summary"
if [[ -d "$results_dir" ]] && [[ -n "$(find "$results_dir" -mindepth 1 -print -quit)" ]]; then
  fail "make explain-target must not create test report artifacts"
fi

make_rows_output="$(env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=rows)"
assert_contains "$make_rows_output" "module.graphprojection.storage.lifecycle:" "make explain-target row mode"

make_target_plan_json="$tmp_dir/make-target-plan.json"
env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET \
  "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" target-plan-json >"$make_target_plan_json"
"$NODE_HELPER" -e 'const plan = JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8")); if (!Array.isArray(plan) || plan.length === 0) process.exit(1);' "$make_target_plan_json"
fi


printf 'target plan smoke passed\n'
