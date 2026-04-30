#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
TASK_GUIDE="$ROOT_DIR/scripts/print-task-guide.mjs"
EXPLAIN_PHASE="$ROOT_DIR/scripts/print-explain-phase.mjs"
EXPLAIN_TARGET="$ROOT_DIR/scripts/print-explain-target.mjs"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/task-guidance.XXXXXX")"
cleanup_paths+=("$tmp_dir")
results_dir="$tmp_dir/results"
mkdir -p "$results_dir/run-a/backend-store"
printf '{"target":"backend-store","status":"pass"}\n' >"$results_dir/run-a/backend-store/target-summary.json"
mkdir -p "$results_dir/run-b" "$results_dir/run-c" "$results_dir/run-d/frontend-unit" "$results_dir/run-e/frontend-unit"
printf '{"label":"check","status":"pass"}\n' >"$results_dir/run-b/run-summary.json"
printf '{"label":"ci","status":"pass"}\n' >"$results_dir/run-c/run-summary.json"
printf '{"target":"not-frontend-unit","status":"pass"}\n' >"$results_dir/run-d/frontend-unit/target-summary.json"
printf '{"target":"frontend-unit","status":"pass"}\n' >"$results_dir/run-e/frontend-unit/target-summary.json"
expected_results_files="$(find "$results_dir" -type f | wc -l | tr -d '[:space:]')"

default_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE")"
assert_contains "$default_output" "Cartulary task guide" "default task-guide header"
assert_contains "$default_output" "local-dev:" "default task-guide local-dev role"
assert_contains "$default_output" "feature-dev:" "default task-guide feature-dev role"
assert_contains "$default_output" "latest_artifact=none" "default task-guide reports missing artifacts"

for role in local-dev feature-dev phase-author ci-investigator release; do
  role_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE" --role "$role")"
  assert_contains "$role_output" "role=$role" "task-guide role header"
  assert_contains "$role_output" "$role:" "task-guide role section"
done

phase_role_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE" --role phase-author --phase phase1)"
assert_contains "$phase_role_output" "phase focus: phase1" "task-guide phase focus"
assert_contains "$phase_role_output" "make explain-phase PHASE=phase1" "task-guide phase command"

guide_json="$tmp_dir/task-guide.json"
"$NODE_BIN" "$TASK_GUIDE" --role feature-dev --json >"$guide_json"
"$NODE_BIN" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$guide_json"

phase4_feature_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE" --role feature-dev --phase phase4)"
assert_contains "$phase4_feature_output" "minimal phase slice: direct targets that cover phase4" "phase4 feature-dev minimal tier"
assert_contains "$phase4_feature_output" "service-backed slice: service-backed targets that cover phase4" "phase4 feature-dev service tier"
assert_contains "$phase4_feature_output" "general hygiene: useful non-phase checks" "phase4 feature-dev hygiene tier"
assert_contains "$phase4_feature_output" "make backend-unit | selected phase execution dependency | phase_relevance=phase_slice" "phase4 backend-unit phase slice"
assert_contains "$phase4_feature_output" "make backend-store | selected phase execution dependency | phase_relevance=phase_slice" "phase4 backend-store phase slice"
assert_contains "$phase4_feature_output" "make backend-integration | selected phase execution dependency | phase_relevance=phase_slice" "phase4 backend-integration phase slice"
assert_contains "$phase4_feature_output" "make backend-integration-support | selected phase execution dependency | phase_relevance=phase_slice" "phase4 backend-integration-support phase slice"
assert_contains "$phase4_feature_output" "make browser-e2e-webserver-backed | selected phase execution dependency | phase_relevance=phase_slice" "phase4 browser phase slice"
assert_contains "$phase4_feature_output" "make frontend-unit | general hygiene outside the selected phase slice | phase_relevance=general_hygiene" "phase4 frontend-unit hygiene"
assert_contains "$phase4_feature_output" "make lint | general hygiene outside the selected phase slice | phase_relevance=general_hygiene" "phase4 lint hygiene"
assert_not_contains "$phase4_feature_output" "make frontend-unit | selected phase execution dependency | phase_relevance=phase_slice" "phase4 frontend-unit not phase slice"
assert_not_contains "$phase4_feature_output" "make lint | selected phase execution dependency | phase_relevance=phase_slice" "phase4 lint not phase slice"

phase4_guide_json="$tmp_dir/task-guide-phase4.json"
"$NODE_BIN" "$TASK_GUIDE" --role feature-dev --phase phase4 --json >"$phase4_guide_json"
"$NODE_BIN" - "$phase4_guide_json" <<'EOF'
const fs = require("node:fs");
const guide = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const role = guide.roles.find((entry) => entry.role === "feature-dev");
if (!role || !Array.isArray(role.recommendation_tiers)) {
  throw new Error("feature-dev must expose recommendation_tiers");
}
if ("recommendations" in role) {
  throw new Error("legacy flat recommendations must not be present");
}
const tierByName = new Map(role.recommendation_tiers.map((tier) => [tier.name, tier]));
for (const name of ["minimal phase slice", "service-backed slice", "full local gate", "general hygiene"]) {
  if (!tierByName.has(name)) {
    throw new Error(`missing tier ${name}`);
  }
}
const minimalTargets = new Set(tierByName.get("minimal phase slice").recommendations.map((item) => item.target));
for (const target of [
  "backend-unit",
  "backend-store",
  "backend-integration",
  "backend-integration-support",
  "browser-e2e-webserver-backed",
]) {
  if (!minimalTargets.has(target)) {
    throw new Error(`phase4 minimal slice missing ${target}`);
  }
}
for (const target of ["frontend-unit", "lint"]) {
  if (minimalTargets.has(target)) {
    throw new Error(`phase4 minimal slice must not include ${target}`);
  }
}
for (const item of tierByName.get("minimal phase slice").recommendations) {
  if (item.phase_relevance !== "phase_slice") {
    throw new Error(`minimal slice item ${item.target} has phase_relevance=${item.phase_relevance}`);
  }
}
const serviceTargets = new Set(tierByName.get("service-backed slice").recommendations.map((item) => item.target));
for (const target of [
  "backend-store",
  "backend-integration",
  "backend-integration-support",
  "browser-e2e-webserver-backed",
]) {
  if (!serviceTargets.has(target)) {
    throw new Error(`phase4 service-backed slice missing ${target}`);
  }
}
if (serviceTargets.has("backend-unit")) {
  throw new Error("backend-unit must not be in the service-backed slice");
}
const gateTargets = new Set(tierByName.get("full local gate").recommendations.map((item) => item.target));
if (!gateTargets.has("test-fast") || !gateTargets.has("check")) {
  throw new Error("full local gate must include test-fast and check");
}
const hygieneTargets = new Set(tierByName.get("general hygiene").recommendations.map((item) => item.target));
if (!hygieneTargets.has("frontend-unit") || !hygieneTargets.has("lint")) {
  throw new Error("phase4 hygiene must include frontend-unit and lint");
}
for (const item of tierByName.get("general hygiene").recommendations) {
  if (item.phase_relevance !== "general_hygiene") {
    throw new Error(`hygiene item ${item.target} has phase_relevance=${item.phase_relevance}`);
  }
}
EOF

phase3_guide_json="$tmp_dir/task-guide-phase3.json"
"$NODE_BIN" "$TASK_GUIDE" --role feature-dev --phase phase3 --json >"$phase3_guide_json"
"$NODE_BIN" - "$phase3_guide_json" <<'EOF'
const fs = require("node:fs");
const guide = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const role = guide.roles.find((entry) => entry.role === "feature-dev");
const tierByName = new Map(role.recommendation_tiers.map((tier) => [tier.name, tier]));
const minimalTargets = new Set(tierByName.get("minimal phase slice").recommendations.map((item) => item.target));
if (!minimalTargets.has("frontend-unit")) {
  throw new Error("phase3 frontend-unit evidence must stay in the minimal phase slice");
}
const hygieneTargets = new Set(tierByName.get("general hygiene").recommendations.map((item) => item.target));
if (hygieneTargets.has("frontend-unit")) {
  throw new Error("phase3 frontend-unit must not be general hygiene");
}
EOF

phase_output="$("$NODE_BIN" "$EXPLAIN_PHASE" --phase phase1)"
assert_contains "$phase_output" "Cartulary phase guidance: phase1" "explain-phase header"
assert_contains "$phase_output" "targets:" "explain-phase targets"
assert_contains "$phase_output" "make backend-store" "explain-phase backend-store target"
assert_contains "$phase_output" "ledger: docs/testing/phase1_coverage_ledger.md" "explain-phase ledger"

phase_json="$tmp_dir/explain-phase.json"
"$NODE_BIN" "$EXPLAIN_PHASE" --phase phase1 --json >"$phase_json"
"$NODE_BIN" - "$phase_json" <<'EOF'
const fs = require("node:fs");
const phase = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (phase.phase !== "phase1" || !Array.isArray(phase.targets) || phase.targets.length === 0) {
  process.exit(1);
}
EOF

unknown_phase_output="$(assert_fails "unknown phase" "$NODE_BIN" "$EXPLAIN_PHASE" --phase phase99)"
assert_contains "$unknown_phase_output" "unknown phase phase99" "unknown phase error"

backend_summary="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target backend-store)"
assert_contains "$backend_summary" "Cartulary target guidance: backend-store" "backend-store explain-target header"
assert_contains "$backend_summary" "services: Postgres,MinIO" "backend-store service requirements"
assert_contains "$backend_summary" "latest_artifact: tmp/task-guidance" "backend-store latest artifact"

backend_rows="$("$NODE_BIN" "$EXPLAIN_TARGET" --target backend-store --detail rows)"
assert_contains "$backend_rows" "rows:" "backend-store rows"
assert_contains "$backend_rows" "U-1-05" "backend-store Go rows"

frontend_output="$("$NODE_BIN" "$EXPLAIN_TARGET" --target frontend-unit)"
assert_contains "$frontend_output" "Cartulary target guidance: frontend-unit" "frontend-unit explain-target"
assert_contains "$frontend_output" "phase_coverage:" "frontend-unit phase coverage"

frontend_identity_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target frontend-unit)"
assert_contains "$frontend_identity_output" "latest_artifact: tmp/task-guidance" "matching target summary latest artifact"
assert_not_contains "$frontend_identity_output" "run-d/frontend-unit/target-summary.json" "mismatched target summary ignored"

test_fast_mismatch_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target test-fast)"
assert_contains "$test_fast_mismatch_output" "latest_artifact: none" "mismatched check run summary ignored for test-fast"

release_mismatch_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target release-check)"
assert_contains "$release_mismatch_output" "latest_artifact: none" "mismatched check run summary ignored for release-check"

ci_match_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target ci)"
assert_contains "$ci_match_output" "run-c/run-summary.json" "matching ci run summary accepted"

browser_output="$("$NODE_BIN" "$EXPLAIN_TARGET" --target browser-e2e-webserver-backed)"
assert_contains "$browser_output" "browser stack" "browser explain-target service requirements"
assert_contains "$browser_output" "webserver-backed browser batch" "browser explain-target scheduler"

check_output="$("$NODE_BIN" "$EXPLAIN_TARGET" --target check)"
assert_contains "$check_output" "check scheduler" "check explain-target scheduler"
assert_contains "$check_output" "phase_coverage:" "check explain-target phase coverage"

target_artifacts="$("$NODE_BIN" "$EXPLAIN_TARGET" --target backend-store --detail artifacts)"
assert_contains "$target_artifacts" "expected:" "target artifact expected paths"
assert_contains "$target_artifacts" "<run-id>/backend-store/target-summary.json" "target artifact expected summary"

target_json="$tmp_dir/explain-target.json"
"$NODE_BIN" "$EXPLAIN_TARGET" --target browser-e2e-webserver-backed --json >"$target_json"
"$NODE_BIN" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$target_json"

unknown_target_output="$(assert_fails "unknown target" "$NODE_BIN" "$EXPLAIN_TARGET" --target no-such-target)"
assert_contains "$unknown_target_output" "unknown target no-such-target" "unknown target error"

make_task_guide="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" task-guide ROLE=feature-dev)"
assert_contains "$make_task_guide" "role=feature-dev" "make task-guide role"

make_phase="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-phase PHASE=phase1)"
assert_contains "$make_phase" "Cartulary phase guidance: phase1" "make explain-phase"

make_target="$(
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=rows
)"
assert_contains "$make_target" "U-1-05" "make explain-target rows"

detail_zero_output="$(assert_fails "obsolete DETAIL=0" "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=0)"
assert_contains "$detail_zero_output" "usage: print-explain-target.mjs" "DETAIL=0 rejected"

if [[ -d "$results_dir" ]] && [[ "$(find "$results_dir" -type f | wc -l | tr -d '[:space:]')" != "$expected_results_files" ]]; then
  fail "guidance commands must not create test report artifacts"
fi
