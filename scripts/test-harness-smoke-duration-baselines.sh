#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
SCRIPT="$ROOT_DIR/scripts/harness-smoke-durations.mjs"

# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

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

assert_fails_with() {
  local label="$1"
  local needle="$2"
  shift 2

  set +e
  local output
  output="$("$@" 2>&1)"
  local status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected command to fail"
  fi
  assert_contains "$output" "$needle" "$label"
}

phase_stdout_from_result() {
  local output="$1"
  local root
  root="$(printf '%s\n' "$output" | sed -n 's/.* run_root=\([^ ]*\) .*/\1/p' | head -n 1)"
  if [[ -z "$root" ]]; then
    fail "missing run_root in output: $output"
  fi
  if [[ "$root" = /* ]]; then
    local dir="$root"
  else
    local dir="$ROOT_DIR/$root"
  fi
  local target
  target="$(printf '%s\n' "$output" | sed -n 's/.* target=\([^ ]*\) .*/\1/p' | head -n 1)"
  [[ -f "$dir/stdout.log" ]] && cat "$dir/stdout.log"
  if [[ -n "$target" ]]; then
    [[ -f "$dir/$target/stdout.log" ]] && cat "$dir/$target/stdout.log"
  fi
}

tmp_dir="$(cartulary_harness_mktemp_dir "harness-smoke-duration-baselines.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

manifest="$tmp_dir/task_surface_manifest.json"
results_dir="$tmp_dir/results"
mkdir -p \
  "$results_dir/harness-smoke-alpha" \
  "$results_dir/harness-smoke-beta" \
  "$results_dir/harness-smoke-gamma" \
  "$results_dir/harness-smoke-failed"

"$NODE_BIN" - "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest" "$ROOT_DIR" "$0" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [source, output, rootDir, backingScript] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
const checks = [
  ["harness-smoke-alpha", "public_make_wrapper"],
  ["harness-smoke-beta", "check_scheduler_semantic"],
  ["harness-smoke-gamma", "service_backed_scheduler_semantic"],
];
for (const check of manifest.harness_checks) {
  delete check.gate_smoke_role;
}
for (const [name, gateRole] of checks) {
  manifest.harness_checks.push({
    name,
    gate_smoke_role: gateRole,
    backing_scripts: [path.relative(rootDir, backingScript).replaceAll("\\", "/")],
  });
}
manifest.harness_tiers.fast = { checks: checks.map(([name]) => name) };
fs.writeFileSync(output, `${JSON.stringify(manifest, null, 2)}\n`);
EOF

cat >"$results_dir/harness-smoke-alpha/target-summary.json" <<'JSON'
{
  "target": "harness-smoke-alpha",
  "status": "pass",
  "wall_duration_ms": 1200
}
JSON
cat >"$results_dir/harness-smoke-beta/target-summary.json" <<'JSON'
{
  "target": "harness-smoke-beta",
  "status": "pass",
  "critical_path_wall_duration_ms": 8400,
  "wall_duration_ms": 99999
}
JSON
cat >"$results_dir/harness-smoke-gamma/target-summary.json" <<'JSON'
{
  "target": "harness-smoke-gamma",
  "status": "pass",
  "logical_duration_ms": 3600
}
JSON
cat >"$results_dir/harness-smoke-failed/target-summary.json" <<'JSON'
{
  "target": "harness-smoke-alpha",
  "status": "fail",
  "wall_duration_ms": 99999
}
JSON

cat >"$tmp_dir/baseline.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {}
}
JSON

update_output="$("$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/baseline.json" --manifest "$manifest" "$results_dir" 2>&1)"
assert_contains "$update_output" "updated 3 harness smoke duration baselines" "baseline update output"

assert_fails_with \
  "missing baseline flag value shows usage" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" update --baseline-file
assert_fails_with \
  "service-only topology flag is rejected" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --manifest "$manifest" --topology "$tmp_dir/topology.json" "$results_dir"
assert_fails_with \
  "multiple harness results dirs are rejected" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/baseline.json" --manifest "$manifest" "$results_dir" "$results_dir"
assert_fails_with \
  "duplicate manifest flags are rejected" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --manifest "$manifest" --manifest "$manifest" "$results_dir"

"$NODE_BIN" - "$tmp_dir/baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
if (baseline.schema_id !== "cartulary.harness_smoke_duration_baselines.v1") {
  throw new Error(`unexpected schema ${baseline.schema_id}`);
}
if (baseline.targets["harness-smoke-alpha"] !== 1200) {
  throw new Error(`expected alpha duration 1200, got ${baseline.targets["harness-smoke-alpha"]}`);
}
if (baseline.targets["harness-smoke-beta"] !== 8400) {
  throw new Error(`expected beta critical path duration 8400, got ${baseline.targets["harness-smoke-beta"]}`);
}
if (baseline.targets["harness-smoke-gamma"] !== 3600) {
  throw new Error(`expected gamma logical duration 3600, got ${baseline.targets["harness-smoke-gamma"]}`);
}
if (Object.hasOwn(baseline.targets, "harness-smoke-failed")) {
  throw new Error("failed harness summaries must not be collected");
}
EOF

"$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --manifest "$manifest" "$results_dir" >/dev/null

cat >"$tmp_dir/missing-observed.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {}
}
JSON
set +e
missing_observed_output="$("$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/missing-observed.json" --manifest "$manifest" "$results_dir/harness-smoke-alpha" 2>&1)"
missing_observed_status=$?
set -e
if [[ "$missing_observed_status" -eq 0 ]]; then
  fail "missing observed summaries should fail baseline update"
fi
assert_contains "$missing_observed_output" "missing observed fast-tier harness target summaries: harness-smoke-beta, harness-smoke-gamma" "missing observed summaries output"

cat >"$tmp_dir/retired.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {
    "harness-smoke-alpha": 1200,
    "harness-smoke-beta": 8400,
    "harness-smoke-gamma": 3600,
    "harness-smoke-retired": 1000
  }
}
JSON
set +e
retired_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/retired.json" --manifest "$manifest" "$results_dir" 2>&1)"
retired_status=$?
set -e
if [[ "$retired_status" -eq 0 ]]; then
  fail "retired baseline drift should fail"
fi
assert_contains "$retired_output" "retired baseline target=harness-smoke-retired" "retired baseline drift output"

cat >"$tmp_dir/missing-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {
    "harness-smoke-alpha": 1200,
    "harness-smoke-beta": 8400
  }
}
JSON
set +e
missing_baseline_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/missing-baseline.json" --manifest "$manifest" "$results_dir" 2>&1)"
missing_baseline_status=$?
set -e
if [[ "$missing_baseline_status" -eq 0 ]]; then
  fail "missing baseline drift should fail"
fi
assert_contains "$missing_baseline_output" "missing harness smoke baseline target=harness-smoke-gamma" "missing baseline drift output"

cat >"$tmp_dir/tolerated-underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {
    "harness-smoke-alpha": 1200,
    "harness-smoke-beta": 3300,
    "harness-smoke-gamma": 3600
  }
}
JSON
"$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/tolerated-underplanned.json" --manifest "$manifest" "$results_dir" >/dev/null

underplanned_results="$tmp_dir/underplanned-results"
cp -R "$results_dir" "$underplanned_results"
cat >"$underplanned_results/harness-smoke-beta/target-summary.json" <<'JSON'
{
  "target": "harness-smoke-beta",
  "status": "pass",
  "critical_path_wall_duration_ms": 40000
}
JSON

cat >"$tmp_dir/underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {
    "harness-smoke-alpha": 1200,
    "harness-smoke-beta": 100,
    "harness-smoke-gamma": 3600
  }
}
JSON
set +e
underplanned_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/underplanned.json" --manifest "$manifest" "$underplanned_results" 2>&1)"
underplanned_status=$?
set -e
if [[ "$underplanned_status" -eq 0 ]]; then
  fail "underplanned baseline drift should fail"
fi
assert_contains "$underplanned_output" "underplanned target=harness-smoke-beta" "underplanned baseline drift output"

cat >"$tmp_dir/overplanned.json" <<'JSON'
{
  "schema_id": "cartulary.harness_smoke_duration_baselines.v1",
  "targets": {
    "harness-smoke-alpha": 1200,
    "harness-smoke-beta": 8400,
    "harness-smoke-gamma": 50000
  }
}
JSON
set +e
overplanned_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/overplanned.json" --manifest "$manifest" "$results_dir" 2>&1)"
overplanned_status=$?
set -e
if [[ "$overplanned_status" -eq 0 ]]; then
  fail "overplanned baseline drift should fail"
fi
assert_contains "$overplanned_output" "overplanned target=harness-smoke-gamma" "overplanned baseline drift output"
